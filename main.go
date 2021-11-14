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
	"sync/atomic"
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
	url string // to solve a problem with relative links on web-page
	doc *goquery.Document
}

// NewPage reads web-page's body
func NewPage(raw io.Reader, url string) (*page, error) {
	doc, err := goquery.NewDocumentFromReader(raw)
	if err != nil {
		return nil, err
	}
	return &page{
		doc: doc,
		url: url,
	}, nil
}

// GetTitle gets title of the 'page'
func (p *page) GetTitle() string {
	return p.doc.Find("title").First().Text()
}

// GetLinks collects a list of links found on the given 'page'
func (p *page) GetLinks() []string {
	var urls []string
	p.doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		url, ok := s.Attr("href")
		prefix := "http"
		if ok {
			// a relative link
			if !strings.HasPrefix(url, prefix) && len(url) > 2 {
				if url[:2] == "//" {
					// add 'http' prefix to the link with '//' at the begining
					url = "http:" + url
				} else if url[:1] == "/" {
					// add page's url to the link with '/' at the begining
					url = p.url + url
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
	Get(ctx context.Context, url string) (Page, error)
}

type requester struct {
	timeout time.Duration
}

func NewRequester(timeout time.Duration) requester {
	return requester{timeout: timeout}
}

// Get searches and returns a webpage on a given URL
func (r requester) Get(ctx context.Context, url string) (Page, error) {
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
		page, err := NewPage(body.Body, url)
		if err != nil {
			return nil, err
		}
		return page, nil
	}
	// return nil, nil     // unreachable code
}

//Crawler - интерфейс (контракт) краулера
type Crawler interface {
	Scan(ctx context.Context, url string, depth uint64)
	ChanResult() <-chan CrawlResult
	IncreaseDepth()
}

// crawler is a main structure that has all the items to control whole process
type crawler struct {
	r        Requester           // a thing that queries pages
	res      chan CrawlResult    // a channel to pass results from r
	visited  map[string]struct{} // a map to hold visited URLs
	mu       sync.RWMutex        // a mutex to share "visited"-map between multibple go-routines
	maxDepth uint64              // limits scanning depth
}

func NewCrawler(r Requester, maxDepth uint64) *crawler {
	return &crawler{
		r:        r,
		res:      make(chan CrawlResult),
		visited:  make(map[string]struct{}),
		mu:       sync.RWMutex{},
		maxDepth: maxDepth,
	}
}

// Scan fills crawler's map with visited URLs and calls Get-method to scan webpages
func (c *crawler) Scan(ctx context.Context, url string, depth uint64) {
	//Проверяем то, что есть запас по глубине
	c.mu.RLock()
	maxDepthAchieved := depth >= c.maxDepth
	c.mu.RUnlock()
	if maxDepthAchieved {
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
		page, err := c.r.Get(ctx, url) //Запрашиваем страницу через Requester
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
			go c.Scan(ctx, link, depth+1) //На все полученные ссылки запускаем новую рутину сборки
		}
	}
}

func (c *crawler) ChanResult() <-chan CrawlResult {
	return c.res
}

// IncreaseDpeth adds 2 to the property 'maxDepth' atomically
func (c *crawler) IncreaseDepth() {
	newDepth := atomic.AddUint64(&c.maxDepth, 2)
	log.Printf("MaxDepth increased via SIGUSR1, new value is %d", newDepth)
}

//Config - структура для конфигурации
type Config struct {
	MaxDepth     uint64 `yaml:"maxdepth"`
	MaxResults   int    `yaml:"maxresults"`
	MaxErrors    int    `yaml:"maxerrors"`
	Url          string `yaml:"url"`
	ReqTimeout   int    `yaml:"reqtimeout"`
	CrawlTimeout int    `yaml:"crawltimeout"`
}

// ReadConfig implements filling config from yaml-file
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
	cr = NewCrawler(r, cfg.MaxDepth)
	log.Printf("Crawler started with PID: %d", os.Getpid())

	ctx, cancel := context.WithCancel(context.Background())
	go cr.Scan(ctx, cfg.Url, 1)             //Запускаем краулер в отдельной рутине
	go processResult(ctx, cancel, cr, *cfg) //Обрабатываем результаты в отдельной рутине

	sigInt := make(chan os.Signal)        //Создаем канал для приема сигналов
	signal.Notify(sigInt, syscall.SIGINT) //Подписываемся на сигнал SIGINT

	sigUsr := make(chan os.Signal)         //Создаем канал для приема сигналов
	signal.Notify(sigUsr, syscall.SIGUSR1) //Подписываемся на сигнал SIGUSR1

	for {
		select {
		case <-ctx.Done(): //Если всё завершили - выходим
			return

		// got INT signal
		case <-sigInt:
			log.Println("got INTERRUPT signal")
			cancel() //Если пришёл сигнал SigInt - завершаем контекст

		// total timeout
		case <-time.After(time.Second * time.Duration(cfg.CrawlTimeout)):
			log.Printf("Crawler stops on timeout: %d sec", cfg.CrawlTimeout)
			cancel()

		// add 2 to max depth
		case <-sigUsr:
			log.Println("got USR1 signal")
			cr.IncreaseDepth() // sigUsr1 - increase maxDepth
		}
	}
}

func processResult(ctx context.Context, cancel func(), cr Crawler, cfg Config) {
	var maxResult, maxErrors = cfg.MaxResults, cfg.MaxErrors
	for {
		select {
		case <-ctx.Done():
			return

		// got message in the channel
		case msg := <-cr.ChanResult():
			if msg.Err != nil {
				// message contains error
				maxErrors--
				log.Printf("crawler result return err: %s\n", msg.Err.Error())
				if maxErrors <= 0 {
					log.Println("Maximum number of errors occured.")
					cancel()
					return
				}
			} else {
				// message contains data
				maxResult--
				log.Printf("crawler result: [url: %s] Title: %s\n", msg.Url, msg.Title)
				if maxResult <= 0 {
					log.Println("Maximum number of results obtained.")
					cancel()
					return
				}
			}
		}
	}
}
