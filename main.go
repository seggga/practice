package main

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/yaml.v3"
)

// CrawlResult is a structure that represents certain status on given page
type CrawlResult struct {
	Err   error
	Title string
	Url   string
}

type Page interface {
	GetTitle() string   // title of the 'page'
	GetLinks() []string // collects a list of links found on the given 'page'
}

// page holds a webpage
type page struct {
	doc *goquery.Document
}

// NewPage reads web-page's body
func NewPage(raw io.Reader) (*page, error) {
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil {
		return nil, err
	}
	return &page{doc: doc}, nil
}

// GetTitle  fit Page interface
func (p *page) GetTitle() string {
	return p.doc.Find("title").First().Text()
}

// GetLinks  fit Page interface
func (p *page) GetLinks() []string {
	var urls []string
	p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		prefix := "http"
		if ok {
			// a trick to solve a problem with relative links - is to add http: at the begining
			if !strings.HasPrefix(url, prefix) {
				if len(url) > 2 && url[:2] == "//" {
					url = "http:" + url
				} else if len(url) > 2 && url[:1] == "/" {
					url = "http:/" + url
				} else {
					return
				}
			}
			urls = append(urls, url)
		}
	})
	return urls
}

type Requester interface {
	GetPage(ctx context.Context, url string) (Page, error)
}

type requester struct {
	timeout time.Duration
}

func NewRequester(timeout time.Duration) requester {
	return requester{timeout: timeout}
}

// GetPage searches and returns a webpage on a given URL
func (r requester) GetPage(ctx context.Context, url string) (Page, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	default:
		cl := &http.Client{
			Timeout: r.timeout,
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		body, err := cl.Do(req)
		if err != nil {
			return nil, err
		}
		defer body.Body.Close()
		page, err := NewPage(body.Body)
		if err != nil {
			return nil, err
		}
		return page, nil
	}
	// return nil, nil     // unreachable code
}

//Crawler - интерфейс (контракт) краулера
type Crawler interface {
	Scan(ctx context.Context, url string, depth int)
	ChanResult() <-chan CrawlResult
}

// crawler is a main structure that has all the items to control whole process
type crawler struct {
	r       Requester           // a thing that queries pages
	res     chan CrawlResult    // a channel to pass results from r
	visited map[string]struct{} // a map to hold visited URLs
	mu      sync.RWMutex        // a mutex to share "visited"-map between multibple go-routines
}

func NewCrawler(r Requester) *crawler {
	return &crawler{
		r:       r,
		res:     make(chan CrawlResult),
		visited: make(map[string]struct{}),
		mu:      sync.RWMutex{},
	}
}

// Scan fills crawler's map with visited URLs and calls GetPage-method to scan webpages
func (c *crawler) Scan(ctx context.Context, url string, depth int) {
	if depth <= 0 { //Проверяем то, что есть запас по глубине
		return
	}
	c.mu.RLock()
	_, ok := c.visited[url] //Проверяем, что мы ещё не смотрели эту страницу
	c.mu.RUnlock()
	if ok {
		return
	}
	select {
	case <-ctx.Done(): //Если контекст завершен - прекращаем выполнение
		return
	default:
		page, err := c.r.GetPage(ctx, url) //Запрашиваем страницу через Requester
		if err != nil {
			c.res <- CrawlResult{Err: err} //Записываем ошибку в канал
			return
		}
		c.mu.Lock()
		c.visited[url] = struct{}{} //Помечаем страницу просмотренной
		c.mu.Unlock()
		c.res <- CrawlResult{ //Отправляем результаты в канал
			Title: page.GetTitle(),
			Url:   url,
		}
		for _, link := range page.GetLinks() {
			go c.Scan(ctx, link, depth-1) //На все полученные ссылки запускаем новую рутину сборки
		}
	}
}

func (c *crawler) ChanResult() <-chan CrawlResult {
	return c.res
}

//Config - структура для конфигурации
type Config struct {
	MaxDepth     int    `yaml:"maxdepth"`
	MaxResults   int    `yaml:"maxresults"`
	MaxErrors    int    `yaml:"maxerrors"`
	Url          string `yaml:"url"`
	ReqTimeout   int    `yaml:"reqtimeout"`
	CrawlTimeout int    `yaml:"crawltimeout"`
}

// Read implements filling config from yaml-file
func ReadConfig() (*Config, error) {
	// read config file
	configData, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	// decode config
	cfg := new(Config)
	err = yaml.Unmarshal(configData, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	// read config file
	cfg, err := ReadConfig()
	if err != nil {
		log.Printf("could not read config file: %v", err)
		return
	}

	var cr Crawler
	var r Requester

	r = NewRequester(time.Duration(cfg.ReqTimeout) * time.Second)
	cr = NewCrawler(r)

	ctx, cancel := context.WithCancel(context.Background())
	go cr.Scan(ctx, cfg.Url, cfg.MaxDepth)  //Запускаем краулер в отдельной рутине
	go processResult(ctx, cancel, cr, *cfg) //Обрабатываем результаты в отдельной рутине

	sigCh := make(chan os.Signal, 2)                      //Создаем канал для приема сигналов
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGUSR1) //Подписываемся на сигнал SIGINT
	for {
		select {
		case <-ctx.Done(): //Если всё завершили - выходим
			return
		case gotSignal := <-sigCh:
			if gotSignal == syscall.SIGINT {
				cancel() //Если пришёл сигнал SigInt - завершаем контекст
			} else if gotSignal == syscall.SIGUSR1 {
				// TODO: добавить глубину +2
			}

		}
	}
}

func processResult(ctx context.Context, cancel func(), cr Crawler, cfg Config) {
	var maxResult, maxErrors = cfg.MaxResults, cfg.MaxErrors
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-cr.ChanResult():
			if msg.Err != nil {
				maxErrors--
				log.Printf("crawler result return err: %s\n", msg.Err.Error())
				if maxErrors <= 0 {
					cancel()
					return
				}
			} else {
				maxResult--
				log.Printf("crawler result: [url: %s] Title: %s\n", msg.Url, msg.Title)
				if maxResult <= 0 {
					cancel()
					return
				}
			}
		case <-time.After(time.Second * time.Duration(cfg.CrawlTimeout)):
			cancel()
			return
		}
	}
}
