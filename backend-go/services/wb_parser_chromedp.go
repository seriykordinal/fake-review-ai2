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

// jitter добавляет небольшой случайный разброс к паузе
func jitter(base, spread time.Duration) time.Duration {
	return base + time.Duration(rand.Int63n(int64(spread)))
}

func countReviews(ctx context.Context) int {
	var n int
	chromedp.Run(ctx, chromedp.Evaluate(
		`document.querySelectorAll('li.comments__item.feedback').length`, &n,
	))
	return n
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
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

// FetchProductReviewsChromedp собирает отзывы через headless Chrome.
//
// Логика прокрутки основана на поведении WB:
//   - Первая загрузка: 30 отзывов + закреплённые (1–5 шт, без даты)
//   - Каждая следующая подгрузка: от 0 до 30 отзывов
//   - WB ограничивает выдачу ~1000 отзывами
//   - Конец загрузки: последняя подгрузка принесла < 30 отзывов
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
	ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Скрываем флаг автоматизации
	chromedp.Run(ctx, chromedp.Evaluate(
		`Object.defineProperty(navigator, 'webdriver', {get: () => undefined});`, nil,
	))

	// ── 1. Открываем страницу товара ─────────────────────────────────────
	log.Printf("Открываем товар %d", productID)
	if err := chromedp.Run(ctx, chromedp.Navigate(productURL)); err != nil {
		return nil, fmt.Errorf("navigate product page: %w", err)
	}
	if !waitForSelector(ctx, ".product-page", 20*time.Second) {
		return nil, fmt.Errorf("product page not loaded (%d)", productID)
	}
	time.Sleep(jitter(1*time.Second, 500*time.Millisecond))

	// ── 2. Переходим на страницу отзывов ─────────────────────────────────
	feedbacksURL := fmt.Sprintf("https://www.wildberries.ru/catalog/%d/feedbacks", productID)
	// Пробуем найти ссылку прямо со страницы
	var linkFromPage string
	chromedp.Run(ctx, chromedp.Evaluate(`
		(() => { let a = document.querySelector('a[href*="feedbacks"]'); return a ? a.href : ""; })()
	`, &linkFromPage))
	if linkFromPage != "" {
		feedbacksURL = linkFromPage
	}

	log.Printf("Переходим на отзывы: %s", feedbacksURL)
	if err := chromedp.Run(ctx, chromedp.Navigate(feedbacksURL)); err != nil {
		return nil, fmt.Errorf("navigate feedbacks: %w", err)
	}

	// ── 3. Ждём первую порцию отзывов (30 + закреплённые) ────────────────
	if !waitForSelector(ctx, "li.comments__item.feedback", 25*time.Second) {
		return nil, fmt.Errorf("no reviews on feedbacks page (%d)", productID)
	}

	// Небольшая пауза — даём DOM устояться после первой загрузки
	time.Sleep(jitter(800*time.Millisecond, 400*time.Millisecond))

	firstLoad := countReviews(ctx)
	log.Printf("Первая загрузка: %d отзывов", firstLoad)

	// Если первая порция < 30 — больше ничего не будет (маленький товар)
	if firstLoad < 30 {
		log.Printf("Товар имеет < 30 отзывов — прокрутка не нужна")
	} else {
		// ── 4. Цикл подгрузки ────────────────────────────────────────────
		// Каждая итерация: скроллим к триггеру → ждём → смотрим прирост.
		// Если прирост < 30 — это последняя порция, выходим.
		// WB лимитирует ~1000 отзывов, поэтому цикл не может быть бесконечным.
		for {
			before := countReviews(ctx)

			// Скроллим к триггеру подгрузки — instant быстрее smooth
			chromedp.Run(ctx, chromedp.Evaluate(`
				(() => {
					let el = document.querySelector('.product-feedbacks__load');
					if (el) el.scrollIntoView({behavior: 'instant', block: 'end'});
				})()
			`, nil))

			// Ждём появления новых отзывов (до 10 секунд)
			after := before
			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				time.Sleep(400 * time.Millisecond)
				after = countReviews(ctx)
				if after > before {
					break
				}
			}

			added := after - before
			log.Printf("Подгрузка: +%d отзывов (всего %d)", added, after)

			// Пришло < 30 — это последняя порция (или пустая), конец
			if added < 30 {
				log.Printf("Последняя порция (%d шт.) — загрузка завершена", added)
				break
			}

			// Короткая пауза между итерациями
			time.Sleep(jitter(300*time.Millisecond, 200*time.Millisecond))
		}
	}

	// ── 5. Парсим DOM ─────────────────────────────────────────────────────
	log.Printf("Парсим %d отзывов...", countReviews(ctx))

	var raw []map[string]string
	err := chromedp.Run(ctx, chromedp.Evaluate(`
		(() => {
			let items  = document.querySelectorAll('li.comments__item.feedback');
			let seen   = new Set();
			let result = [];

			for (let item of items) {
				// Рейтинг из CSS-класса star1-star5
				let rating = "";
				let starsEl = item.querySelector('.feedback__rating');
				if (starsEl) {
					let m = starsEl.className.match(/star([1-5])/);
					if (m) rating = m[1];
				}

				// Дата — только из блока покупателя (.feedback__info)
				// Закреплённые отзывы даты не имеют — это нормально
				let dateEl = item.querySelector('.feedback__info .feedback__date');
				let date   = dateEl ? dateEl.innerText.trim() : "";

				// Текст — p[itemprop="reviewBody"] есть только у отзыва покупателя,
				// ответ продавца (.feedbackContainer) этого атрибута не имеет
				let pros = "", cons = "", text = "";
				let body = item.querySelector('p[itemprop="reviewBody"]');
				if (body) {
					let spans = body.querySelectorAll('span.feedback__text--item');
					for (let span of spans) {
						let bold    = span.querySelector('.feedback__text--item-bold');
						let label   = bold ? bold.innerText.trim() : "";
						let content = span.innerText.replace(label, "").trim();
						if      (label === "Достоинства:") pros = content;
						else if (label === "Недостатки:")  cons = content;
						else if (label === "Комментарий:") text = content;
						else if (label === "")             text = span.innerText.trim();
					}
					if (!pros && !cons && !text) text = body.innerText.trim();
				}

				// Тег-плюсы из feedbacks-bables (если pros пустой)
				if (!pros) {
					let bables = item.querySelector('.feedbacks-bables');
					if (bables) {
						let tags = Array.from(bables.querySelectorAll('.feedbacks-bables__item'))
							.map(el => el.innerText.trim());
						if (tags.length) pros = tags.join(", ");
					}
				}

				if (!text && !pros && !cons) continue;

				// Дедупликация
				let key = rating + "|" + date + "|" + text + "|" + pros + "|" + cons;
				if (seen.has(key)) continue;
				seen.add(key);

				result.push({ rating, date, pros, cons, text });
			}
			return result;
		})()
	`, &raw))

	if err != nil {
		return nil, fmt.Errorf("parse JS error: %w", err)
	}

	var reviews []models.Review
	for _, r := range raw {
		reviews = append(reviews, models.Review{
			Rating: r["rating"],
			Date:   r["date"],
			Pros:   r["pros"],
			Cons:   r["cons"],
			Text:   r["text"],
		})
	}

	log.Printf("Итого: %d отзывов", len(reviews))

	if len(reviews) == 0 {
		return nil, fmt.Errorf("no reviews found for product %d", productID)
	}
	return reviews, nil
}
