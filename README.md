# AllyMailer Go net/http 1.0.0

Web API 1.90

Покроковий посібник: [Надсилання листів на Go](https://docs.allymailer.com.ua/ua/send-with-go) · Документація для LLM: [llms.txt](https://docs.allymailer.com.ua/llms.txt) · [Довідник API](https://docs.allymailer.com.ua/ua/web-api-ref)

Інтеграція [Go SDK AllyMailer](https://github.com/allymailer/allymailer-go-sdk) (`github.com/allymailer/allymailer-go-sdk`) з Go net/http: клієнт API в request context (net/http middleware), ендпоінт вебхуків із перевіркою підпису.

## Встановлення

```bash
go get github.com/allymailer/allymailer-nethttp
```

Вимоги: Go 1.21+.

## Отримайте API-ключ

1. Зареєструйтеся на [cp.allymailer.com](https://cp.allymailer.com).
2. Скопіюйте API-ключ у [панелі керування](https://cp.allymailer.com).
3. Покладіть його у змінну оточення:

```bash
export ALLYMAILER_API_KEY="your-api-key"
```

Пакет бере його з `ALLYMAILER_API_KEY` (див. налаштування нижче).

> **Лише на сервері.** Використовуйте цей пакет у серверному коді: API-маршрути, серверні функції, воркери.
> Ніколи не передавайте API-ключ у браузер чи мобільний застосунок і не комітьте його.

## Налаштування

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

## Перший лист — без налаштування DNS

Кожен обліковий запис може одразу надсилати тестові листи із **sandbox-домену**: він уже підтверджений, DKIM і SPF налаштовані.
Листи доставляються лише на підтверджені вами адреси.

У прикладах — клієнт API з пакета (див. налаштування вище).

**1. Отримайте sandbox-домен** (`domain/get-sandbox-domain`). Перший виклик створює домен, наступні повертають той самий.

```go
import (
	"fmt"
	"net/http"

	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

mux.HandleFunc("/sandbox-domain", func(w http.ResponseWriter, r *http.Request) {
	client := allymailerwrapper.ClientFromContext(r.Context())
	response, err := client.Domain.GetSandboxDomain(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, response.Domain)
})
```

**2. Підтвердіть свою скриньку** (`confirmed-email/create`). AllyMailer надішле лист із підтвердженням; перейдіть за посиланням із нього протягом 48 годин.

```go
import (
	"fmt"
	"net/http"

	"github.com/allymailer/allymailer-go-sdk/api/confirmedemail"
	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

mux.HandleFunc("/confirm-email", func(w http.ResponseWriter, r *http.Request) {
	client := allymailerwrapper.ClientFromContext(r.Context())
	response, err := client.ConfirmedEmail.Create(r.Context(), &confirmedemail.ConfirmedEmailRequest{
		Email: "you@example.com",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, response.Status)
})
```

Статус перевірте через `confirmed-email/list` — у вашої адреси має бути `"status": "confirmed"`.

**3. Надішліть лист** (`email/send`) з будь-якої адреси sandbox-домену на підтверджену адресу.

```go
import (
	"fmt"
	"net/http"

	allymailer "github.com/allymailer/allymailer-go-sdk"
	"github.com/allymailer/allymailer-go-sdk/api/email"
	"github.com/allymailer/allymailer-go-sdk/common"
	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
	client := allymailerwrapper.ClientFromContext(r.Context())
	response, err := client.Email.Send(r.Context(), &email.EmailSendRequest{
		Message: &email.MessageObject{
			Recipients: []email.MessageObjectRecipientsItem{
				{
					Email: "you@example.com",
				},
			},
			Body: &common.BodyObject{
				HTML: allymailer.String("<p>Працює!</p>"),
			},
			Subject:   "Привіт від AllyMailer",
			FromEmail: allymailer.String("noreply@<sandbox-domain>"),
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, response.JobID)
})
```

### Що повертається

```json
{
  "status": "success",
  "job_id": "1ZymBc-00041N-9X",
  "emails": [
    "you@example.com"
  ]
}
```

У разі помилки API повертає `status: "error"`, `message` і числовий `code`:

```json
{
  "status": "error",
  "code": 102,
  "message": "Error ID:83274F98-C7D5-11F1-B221-6214FF8707B1. Can not decode key wrongkey"
}
```

### Те саме через curl

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

## Надсилання через API

Клієнт викликає [email/send](https://docs.allymailer.com.ua/ua/web-api-ref#email-send) з усіма його параметрами: шаблони, підстановки, метадані, трекінг. Поля запиту та інші приклади — у [документації Go SDK](https://github.com/allymailer/allymailer-go-sdk/blob/HEAD/docs/Email.md#send).

```go
import (
	"fmt"
	"net/http"

	allymailer "github.com/allymailer/allymailer-go-sdk"
	"github.com/allymailer/allymailer-go-sdk/api/email"
	"github.com/allymailer/allymailer-go-sdk/common"
	allymailerwrapper "github.com/allymailer/allymailer-nethttp"
)

mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
	client := allymailerwrapper.ClientFromContext(r.Context())
	response, err := client.Email.Send(r.Context(), &email.EmailSendRequest{
		Message: &email.MessageObject{
			Recipients: []email.MessageObjectRecipientsItem{
				{
					Email: "user@example.com",
					Substitutions: map[string]string{
						"CustomerId": "12452",
						"to_name":    "John Smith",
					},
				},
			},
			Body: &common.BodyObject{
				HTML:      allymailer.String("<b>Hello, {{to_name}}</b>"),
				Plaintext: allymailer.String("Hello, {{to_name}}"),
			},
			Subject:   "AllyMailer test email ",
			FromEmail: allymailer.String("email@example.com"),
			FromName:  allymailer.String("John Smith"),
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	fmt.Fprintln(w, response.JobID)
})
```

## Обмеження sandbox

- Надсилання лише на підтверджені адреси: до 10 на обліковий запис.
- До 100 листів на день.
- Sandbox-домен спливає через 31 день; знову викличте `domain/get-sandbox-domain`, щоб отримати новий.
- Працює лише з API-ключем облікового запису, не з ключем проєкту.
- Не показується в `domain/list`; інші методи `domain/*` до нього не застосовуються.

Sandbox — для розробки та тестування. Не використовуйте sandbox-домен як відправника на проді.

## Перед продом: підтвердіть свій домен

1. `domain/get-dns-records` — повертає значення для DNS-записів.
2. Додайте ці записи у свого DNS-провайдера (їх також показує сторінка домену в панелі керування):

   | Тип | Хост | Значення |
   |---|---|---|
   | TXT | `@` | значення `verification-record` (`allymailer-validate-hash=…`) |
   | TXT | `<селектор>._domainkey` (селектор — на сторінці домену в панелі керування, наприклад `us`) | `v=DKIM1; k=rsa; p=<dkim>` |
   | TXT | `@` | `v=spf1 include:spf.allymailer.com ~all` |
   | CNAME | `_dmarc` | значення `dmarc` |

3. `domain/validate-verification-record`, потім `domain/validate-dkim`.
4. `domain/list` — дочекайтеся, поки запис підтвердження буде підтверджено, а DKIM — активний.
5. Беріть `from_email` з конфігурації, щоб перехід із sandbox не потребував змін у коді.

## Надійне надсилання

- **Повтори:** ставте `message.idempotence_key` у надсиланнях, які можуть повторюватися, щоб повторний запит не надіслав лист двічі.
- **Ліміти:** до 500 отримувачів і 10 МБ на запит; кожне вкладення до 7 МБ.
- **Тіло:** завжди передавайте `body.html` або `body.plaintext`.
- **Статус доставки:** зареєструйте вебхук (`webhook/set`) зі швидким ідемпотентним HTTPS-обробником замість опитування.
- **Типові налаштування:** не прибирайте стандартний блок відписки та трекінг; змінити їх можна через підтримку AllyMailer.
- **Помилки:** повторюйте `429` і `5xx` з backoff; інші помилки без змін не повторюйте.

## Посилання

- [Надсилання листів на Go](https://docs.allymailer.com.ua/ua/send-with-go) та інші [посібники з інтеграції](https://docs.allymailer.com.ua/ua/integrations/)
- [Довідник Web API](https://docs.allymailer.com.ua/ua/web-api-ref)
- [llms.txt](https://docs.allymailer.com.ua/llms.txt) для AI-асистентів
- [Підтримка](https://cp.allymailer.com/ua/support)

## Для розробників пакета

Див. [AGENTS.md](AGENTS.md): як цей пакет генерується з OpenAPI-специфікації, збирається, перевіряється та випускається.
