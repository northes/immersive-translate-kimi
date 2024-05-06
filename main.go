package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/northes/go-moonshot"
)

var (
	moonshotClient     *moonshot.Client
	moonshotClientOnce sync.Once
)

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "3000"
	}

	app := fiber.New(fiber.Config{
		StructValidator: &structValidator{
			validate: validator.New(),
		},
		ErrorHandler: func(c fiber.Ctx, err error) error {
			resp := &Response{
				Translations: make([]*Translation, 0),
			}
			errs := strings.Split(err.Error(), "\n")
			for _, e := range errs {
				resp.Translations = append(resp.Translations, &Translation{
					DetectedSourceLang: e,
				})
			}
			return c.JSON(resp)
		},
	})
	app.Post("/", HandleTranslation)

	log.Fatal(app.Listen(fmt.Sprintf(":%s", port)))
}

type Request struct {
	SourceLang string   `json:"source_lang" validate:"required"`
	TargetLang string   `json:"target_lang" validate:"required"`
	TextList   []string `json:"text_list" validate:"required"`
}

type Response struct {
	Translations []*Translation `json:"translations"`
}

type Translation struct {
	DetectedSourceLang string `json:"detected_source_lang"`
	Text               string `json:"text"`
}

func HandleTranslation(c fiber.Ctx) error {
	key := c.Query("key")
	if moonshotClient == nil {
		moonshotClientOnce.Do(func() {
			moonshotClient, _ = moonshot.NewClient(key)
		})
	}

	req := new(Request)
	if err := c.Bind().JSON(req); err != nil {
		return err
	}

	moonshotResp, err := moonshotClient.Chat().Completions(context.Background(), &moonshot.ChatCompletionsRequest{
		Model:       moonshot.ModelMoonshotV18K,
		Temperature: 0.0,
		Stream:      false,
		Messages: []*moonshot.ChatCompletionsMessage{
			{
				Role:    moonshot.RoleSystem,
				Content: "你是一个专业的翻译员，你需要将用户输入的内容进行并输出。用户输入的是中文则翻译成英文，用户输入的是其他语言则翻译成中文。如果输入的是单词则翻译结果需要小写开头，如果输入的是句子则翻译结果需要注意大小写",
			},
			{
				Role:    moonshot.RoleUser,
				Content: req.TextList[0],
			},
		},
	})
	if err != nil {
		return err
	}

	respContent := moonshotResp.Choices[0].Message.Content

	return c.JSON(&Response{
		Translations: []*Translation{
			{
				DetectedSourceLang: req.SourceLang,
				Text:               respContent,
			},
		},
	})
}
