# AllyMailer Go net/http

[![Go net/http](https://pkg.go.dev/badge/github.com/allymailer/allymailer-nethttp.svg)](https://pkg.go.dev/github.com/allymailer/allymailer-nethttp) [![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

Інтеграція [Go SDK AllyMailer](https://github.com/allymailer/allymailer-go-sdk) (`github.com/allymailer/allymailer-go-sdk`) з Go net/http: клієнт API в request context (net/http middleware), ендпоінт вебхуків із перевіркою підпису.

[Довідник API](https://docs.allymailer.com.ua/ua/web-api-ref) · Документація для LLM: [llms.txt](https://docs.allymailer.com.ua/llms.txt)

## Встановлення

```bash
go get github.com/allymailer/allymailer-nethttp
```

Вимоги: Go 1.21+.

## Налаштування

Отримайте API-ключ у [панелі керування](https://cp.allymailer.com) і зберігайте його у змінній оточення, а не в коді:

```bash
export ALLYMAILER_API_KEY="your-api-key"
```

За замовчуванням налаштування беруться зі змінних оточення:

```bash
ALLYMAILER_API_KEY=YOUR_API_KEY
ALLYMAILER_HOST=cp.allymailer.com
```

| Параметр | Змінна оточення | За замовчуванням | Опис |
|---|---|---|---|
| `APIKey` | `ALLYMAILER_API_KEY` |  | API-ключ користувача або проєкту (обов'язковий). |
| `Host` | `ALLYMAILER_HOST` | `cp.allymailer.com` | Хост API (датацентр: `cp.allymailer.com`) або повний базовий URL. |
| `Timeout` | `ALLYMAILER_TIMEOUT` | `30` | Таймаут запиту, секунди. |
| `WebhookKey` | `ALLYMAILER_WEBHOOK_KEY` |  | API-ключ, яким підписані вебхуки (ключ користувача/проєкту вебхука), якщо він відрізняється від API-ключа. |
| `WebhookPath` | `ALLYMAILER_WEBHOOK_PATH` | `/allymailer/webhook` | Шлях ендпоінта вебхуків. |

`ConfigFromEnv()` читає змінні; `Config` можна заповнити і в коді. `ALLYMAILER_TIMEOUT` задається в секундах, `Config.Timeout` — `time.Duration`.

## Вебхуки

Ендпоінт відповідає 200 OK на `GET` (перевірка URL під час [webhook/set](https://docs.allymailer.com.ua/ua/web-api-ref#webhook-set)) і приймає `POST`-сповіщення: перевіряє поле `auth` (MD5 тіла запиту, у якому значення `auth` замінено на API-ключ), на неправильний підпис відповідає 401, а сповіщення і кожну його подію передає застосунку.

```go
import (
	"context"
	"log"

	"github.com/allymailer/allymailer-go-sdk/api/webhook"
	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

cfg := allymailerwrapper.ConfigFromEnv()
handler := allymailerwrapper.WebhookHandler(cfg.SigningKey(), allymailerwrapper.Handlers{
	OnEvent: func(ctx context.Context, e *webhook.WebhookCallbackEventsByUserItemEventsItem, user *webhook.WebhookCallbackEventsByUserItem) error {
		if e.EventData != nil {
			log.Printf("AllyMailer event %s: %s %s", e.EventName, e.EventData.Email, e.EventData.Status)
		}
		return nil
	},
})

mux.Handle(cfg.Path(), handler)           // net/http, chi: r.Handle(cfg.Path(), handler)
router.Any(cfg.Path(), gin.WrapH(handler)) // gin
e.Any(cfg.Path(), echo.WrapHandler(handler)) // echo
```

Зареєструйте URL ендпоінта методом `webhook/set`, наприклад `https://example.com/allymailer/webhook`.

> **Відповідайте 200 протягом 3 секунд**, а події обробляйте у фоні (через чергу): інакше AllyMailer повторює запит кожні 10 хвилин протягом доби (зростає `retry_count`), а обробник, на якому за добу не пройшов жоден запит хоча б для 10 подій, вимикається. Обробник має витримувати **щонайменше 5 паралельних запитів до 100 подій у кожному**; повтор надсилає ті самі події ще раз — прибирайте дублікати.

## Клієнт API

Усі розділи API доступні через клієнт, див. [документацію SDK](https://github.com/allymailer/allymailer-go-sdk).

```go
import (
	"fmt"
	"net/http"

	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

cfg := allymailerwrapper.ConfigFromEnv()
client := cfg.Client()

mux := http.NewServeMux()
mux.HandleFunc("/usage", func(w http.ResponseWriter, r *http.Request) {
	info, err := allymailerwrapper.ClientFromContext(r.Context()).System.Info(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, info.Email)
})
http.ListenAndServe(":8080", allymailerwrapper.ClientMiddleware(client)(mux))
```

### Надсилання через API

Клієнт викликає [email/send](https://docs.allymailer.com.ua/ua/web-api-ref#email-send) з усіма його параметрами: шаблони, підстановки, метадані, трекінг. Поля запиту та інші приклади — у [документації Go SDK](https://github.com/allymailer/allymailer-go-sdk/blob/HEAD/docs/Email.md#send).

```go
import (
	"net/http"

	allymailer "github.com/allymailer/allymailer-go-sdk"
	"github.com/allymailer/allymailer-go-sdk/api/email"
	"github.com/allymailer/allymailer-go-sdk/common"
	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := allymailerwrapper.ClientFromContext(ctx)

	if _, err := client.Email.Send(ctx, &email.EmailSendRequest{
		Message: &email.MessageObject{
			Recipients: []email.MessageObjectRecipientsItem{email.RecipientFromAddress("Олексій <alex@example.com>")},
			FromEmail:  allymailer.String("shop@example.com"),
			FromName:   allymailer.String("Магазин"),
			Subject:    "Ваше замовлення",
			Body:       common.HTMLBody("<p>Вітаємо, {{to_name}}! Ваше замовлення надіслано.</p>"),
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
})
```

## Без свого домену: тестовий sandbox-домен

Поки ваш домен не підтверджено, листи можна надсилати із **sandbox-домену** — тестового домену, який є в кожного облікового запису: він уже підтверджений, DKIM і SPF налаштовані. Листи з нього доставляються лише на підтверджені вами адреси (посилання для підтвердження приходить листом і діє 48 годин), до 100 на день. Через 31 день після створення sandbox-домен видаляється, але його можна отримати знову: наступний виклик `domain/get-sandbox-domain` створить новий (з іншою назвою, тому беріть її з відповіді, а не прописуйте в коді).

У прикладах — клієнт API з пакета (див. налаштування вище).

```go
import (
	"net/http"

	allymailer "github.com/allymailer/allymailer-go-sdk"
	"github.com/allymailer/allymailer-go-sdk/api/confirmedemail"
	"github.com/allymailer/allymailer-go-sdk/api/email"
	"github.com/allymailer/allymailer-go-sdk/common"
	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

mux.HandleFunc("/sandbox", func(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	client := allymailerwrapper.ClientFromContext(ctx)

	// 1. Ваш sandbox-домен: створюється під час першого виклику, далі повертається той самий
	sandbox, err := client.Domain.GetSandboxDomain(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// 2. Підтвердіть свою скриньку: AllyMailer надішле посилання, перейдіть за ним протягом 48 годин
	confirm := &confirmedemail.ConfirmedEmailRequest{Email: "you@example.com"}
	if _, err := client.ConfirmedEmail.Create(ctx, confirm); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	// 3. Надішліть із будь-якої адреси sandbox-домену на підтверджену адресу
	if _, err := client.Email.Send(ctx, &email.EmailSendRequest{
		Message: &email.MessageObject{
			Recipients: []email.MessageObjectRecipientsItem{email.RecipientFromAddress("you@example.com")},
			FromEmail:  allymailer.String("noreply@" + sandbox.Domain),
			Subject:    "Привіт від AllyMailer",
			Body:       common.HTMLBody("<p>Працює!</p>"),
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
})
```

<details>
<summary>Обмеження sandbox</summary>

- Надсилання лише на підтверджені адреси: до 10 на обліковий запис.
- До 100 листів на день.
- Sandbox-домен спливає через 31 день; знову викличте `domain/get-sandbox-domain`, щоб отримати новий.
- Працює лише з API-ключем облікового запису, не з ключем проєкту.
- Не показується в `domain/list`; інші методи `domain/*` до нього не застосовуються.

Sandbox — для розробки та тестування. Не використовуйте sandbox-домен як відправника на проді.

</details>

<details>
<summary>Те саме через curl</summary>

```bash
BASE=https://cp.allymailer.com/ua/transactional/api/v1

curl -s -X POST $BASE/domain/get-sandbox-domain.json \
  -H "Content-Type: application/json" -H "X-API-KEY: $ALLYMAILER_API_KEY" -d '{}'

curl -s -X POST $BASE/confirmed-email/create.json \
  -H "Content-Type: application/json" -H "X-API-KEY: $ALLYMAILER_API_KEY" \
  -d '{"email": "you@example.com"}'

curl -s -X POST $BASE/email/send.json \
  -H "Content-Type: application/json" -H "X-API-KEY: $ALLYMAILER_API_KEY" \
  -d '{
        "message": {
          "recipients": [
            {
              "email": "you@example.com"
            }
          ],
          "from_email": "noreply@<sandbox-domain>",
          "subject": "Привіт від AllyMailer",
          "body": {
            "html": "<p>Працює!</p>"
          }
        }
      }'
```

</details>

## Запуск у продакшен

**1. Підтвердіть домен надсилання.** `domain/get-dns-records` повертає запис підтвердження, DKIM-ключ і адресу для DMARC; додайте записи у свого DNS-провайдера, потім `domain/validate-verification-record` і `domain/validate-dkim`.

| Тип | Хост | Значення | Звідки взяти |
|---|---|---|---|
| TXT | `@` | `allymailer-validate-hash=…` | поле `verification-record` відповіді, як є |
| TXT | `<селектор>._domainkey` (селектор — на сторінці домену в панелі керування, наприклад `us`) | `v=DKIM1; k=rsa; p=<dkim>` | поле `dkim` відповіді після `v=DKIM1; k=rsa; p=` |
| TXT | `@` | `v=spf1 include:spf.allymailer.com ~all` | у відповіді немає; якщо SPF-запис уже є, додайте include до нього |
| CNAME | `_dmarc` | `<домен>.dmarc.…` | поле `dmarc` відповіді (DMARC-політику публікує AllyMailer); якщо в домену вже є свій TXT-запис DMARC, залиште його |

**2. Надійне надсилання.**

| | |
|---|---|
| Повтори | `429` і `5xx` з backoff і тим самим `idempotence_key`; інші помилки виправляйте, а не повторюйте |
| Ліміти | 500 одержувачів і 10 МБ на запит разом із вкладеннями: у base64 файл стає приблизно на третину більшим |
| Тіло | завжди `html` або `plaintext` (або із шаблону) |
| Статус доставки | вебхук, а не опитування |
| Черга | надсилайте з фонового завдання, а не всередині веб-запиту |
| Налаштування | `from_email` з конфігурації: перехід із sandbox на продакшен без змін у коді |
| Типові налаштування | не прибирайте блок відписки та трекінг; змінити їх можна через підтримку AllyMailer |

## Допомога

- [Go SDK](https://github.com/allymailer/allymailer-go-sdk)
- Для ШІ-агентів: [llms.txt](https://docs.allymailer.com.ua/llms.txt), скіли: `npx skills add https://github.com/allymailer/allymailer-skills.git`
- [Підтримка](https://cp.allymailer.com/ua/support) · [Задачі](https://github.com/allymailer/allymailer-nethttp/issues)
- Розробникам: [AGENTS.md](AGENTS.md) — пакет генерується з OpenAPI-специфікації.
