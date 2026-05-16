package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"fake-review-ai2/models"

	"github.com/chromedp/chromedp"
)

type WBParserService struct{}

func NewWBParserService() *WBParserService {
	return &WBParserService{}
}

func (s *WBParserService) ExtractProductID(url string) (int, error) {
	re := regexp.MustCompile(`/catalog/(\d+)/`)
	match := re.FindStringSubmatch(url)
	if len(match) < 2 {
		return 0, fmt.Errorf("invalid product URL: %s", url)
	}
	return strconv.Atoi(match[1])
}

func humanDelay(base, jitter time.Duration) {
	d := base + time.Duration(rand.Int63n(int64(jitter)))
	time.Sleep(d)
}

func humanScroll(ctx context.Context, pixels int) error {
	step := 80 + rand.Intn(60)
	scrolled := 0
	for scrolled < pixels {
		chunk := step
		if scrolled+chunk > pixels {
			chunk = pixels - scrolled
		}
		js := fmt.Sprintf(`window.scrollBy({top: %d, behavior: 'smooth'});`, chunk)
		if err := chromedp.Run(ctx, chromedp.Evaluate(js, nil)); err != nil {
			return err
		}
		scrolled += chunk
		time.Sleep(time.Duration(40+rand.Intn(60)) * time.Millisecond)
	}
	return nil
}

func waitForSelector(ctx context.Context, sel string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var exists bool
		chromedp.Run(ctx, chromedp.Evaluate(
			fmt.Sprintf(`document.querySelector('%s') !== null`, sel), &exists,
		))
		if exists {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	return false
}

func countReviews(ctx context.Context) int {
	var n int
	chromedp.Run(ctx, chromedp.Evaluate(
		`document.querySelectorAll('li.comments__item.feedback').length`, &n,
	))
	return n
}

// FetchProductReviewsChromedp собирает отзывы с помощью chromedp.
// Алгоритм:
//  1. Открываем страницу товара, ждём полной загрузки
//  2. Переходим в раздел отзывов
//  3. Ждём появления первых отзывов в DOM
//  4. Плавно скроллим к триггеру .product-feedbacks__load пока счётчик растёт
//  5. Парсим отзывы покупателей, исключая ответы продавца
func (s *WBParserService) FetchProductReviewsChromedp(productID int) ([]models.Review, error) {
	productURL := fmt.Sprintf("https://www.wildberries.ru/catalog/%d/detail.aspx", productID)

	chromePath := `C:\Program Files\Google\Chrome\Application\chrome.exe`

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.WindowSize(1366, 768),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Evaluate(`
		Object.defineProperty(navigator, 'webdriver', {get: () => undefined});
	`, nil)); err != nil {
		log.Printf("Warning: не удалось скрыть webdriver: %v", err)
	}

	var reviews []models.Review

	log.Printf("Открываем страницу товара: %s", productURL)
	if err := chromedp.Run(ctx, chromedp.Navigate(productURL)); err != nil {
		return nil, fmt.Errorf("navigate to product page: %w", err)
	}

	if !waitForSelector(ctx, ".product-page", 20*time.Second) {
		return nil, fmt.Errorf("product page did not load (product %d)", productID)
	}
	humanDelay(3*time.Second, 2*time.Second)

	humanScroll(ctx, 300+rand.Intn(200))
	humanDelay(1*time.Second, 1*time.Second)

	feedbacksURL := fmt.Sprintf("https://www.wildberries.ru/catalog/%d/feedbacks", productID)

	var linkFromPage string
	chromedp.Run(ctx, chromedp.Evaluate(`
		(() => {
			let a = document.querySelector('a[href*="feedbacks"]');
			return a ? a.href : "";
		})()
	`, &linkFromPage))
	if linkFromPage != "" {
		feedbacksURL = linkFromPage
	}

	log.Printf("Переходим на отзывы: %s", feedbacksURL)
	if err := chromedp.Run(ctx, chromedp.Navigate(feedbacksURL)); err != nil {
		return nil, fmt.Errorf("navigate to feedbacks: %w", err)
	}

	log.Printf("Ждём загрузки первых отзывов...")
	if !waitForSelector(ctx, "li.comments__item.feedback", 25*time.Second) {
		return nil, fmt.Errorf("no reviews appeared on feedbacks page (product %d)", productID)
	}

	humanDelay(2*time.Second, 2*time.Second)

	firstCount := countReviews(ctx)
	log.Printf("Первые отзывы загружены: %d", firstCount)

	noGrowthStreak := 0
	const maxNoGrowth = 2

	for {
		before := countReviews(ctx)
		log.Printf("Отзывов в DOM: %d", before)

		var hasTrigger bool
		chromedp.Run(ctx, chromedp.Evaluate(
			`document.querySelector('.product-feedbacks__load') !== null`, &hasTrigger,
		))
		if !hasTrigger {
			log.Printf("Триггер исчез из DOM — все отзывы загружены")
			break
		}

		humanScroll(ctx, 400+rand.Intn(300))
		humanDelay(800*time.Millisecond, 600*time.Millisecond)

		chromedp.Run(ctx, chromedp.Evaluate(`
			(() => {
				let el = document.querySelector('.product-feedbacks__load');
				if (el) el.scrollIntoView({behavior: 'smooth', block: 'center'});
			})()
		`, nil))

		after := before
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(500 * time.Millisecond)
			after = countReviews(ctx)
			if after > before {
				break
			}
		}

		if after > before {
			log.Printf("Загружено +%d отзывов (всего %d)", after-before, after)
			noGrowthStreak = 0
			humanDelay(1500*time.Millisecond, 1500*time.Millisecond)
		} else {
			noGrowthStreak++
			log.Printf("Нет новых отзывов (попытка %d/%d)", noGrowthStreak, maxNoGrowth)
			if noGrowthStreak >= maxNoGrowth {
				log.Printf("Загрузка завершена — отзывы перестали появляться")
				break
			}
			humanDelay(3*time.Second, 2*time.Second)
		}
	}

	log.Printf("Начинаем парсинг отзывов...")
	var raw []map[string]string
	err := chromedp.Run(ctx, chromedp.Evaluate(`
		(() => {
			let items = document.querySelectorAll('li.comments__item.feedback');
			let seen  = new Set();
			let result = [];

			for (let item of items) {
				let rating = "";
				let starsEl = item.querySelector('.feedback__rating');
				if (starsEl) {
					let m = starsEl.className.match(/star([1-5])/);
					if (m) rating = m[1];
				}

				let dateEl = item.querySelector('.feedback__info .feedback__date');
				let date   = dateEl ? dateEl.innerText.trim() : "";

				
				let pros = "", cons = "", text = "";
				let reviewBody = item.querySelector('p[itemprop="reviewBody"]');
				if (reviewBody) {
					let spans = reviewBody.querySelectorAll('span.feedback__text--item');
					for (let span of spans) {
						let boldEl  = span.querySelector('.feedback__text--item-bold');
						let label   = boldEl ? boldEl.innerText.trim() : "";
						let content = span.innerText.replace(label, "").trim();

						if      (label === "Достоинства:")  pros = content;
						else if (label === "Недостатки:")   cons = content;
						else if (label === "Комментарий:")  text = content;
						else if (label === "")              text = span.innerText.trim();
					}
					if (!pros && !cons && !text) text = reviewBody.innerText.trim();
				}

				if (!pros) {
					let bables = item.querySelector('.feedbacks-bables');
					if (bables) {
						let tags = Array.from(bables.querySelectorAll('.feedbacks-bables__item'))
							.map(el => el.innerText.trim());
						if (tags.length) pros = tags.join(", ");
					}
				}

				if (!text && !pros && !cons) continue;

				let key = rating + "|" + date + "|" + text + "|" + pros + "|" + cons;
				if (seen.has(key)) continue;
				seen.add(key);

				result.push({ rating, date, pros, cons, text });
			}
			return result;
		})()
	`, &raw))

	if err != nil {
		return nil, fmt.Errorf("parse reviews JS error: %w", err)
	}

	for _, r := range raw {
		reviews = append(reviews, models.Review{
			Rating: r["rating"],
			Date:   r["date"],
			Pros:   r["pros"],
			Cons:   r["cons"],
			Text:   r["text"],
		})
	}

	log.Printf("Итого спарсено отзывов: %d", len(reviews))

	if len(reviews) == 0 {
		return nil, fmt.Errorf("no reviews found for product %d", productID)
	}
	return reviews, nil
}
