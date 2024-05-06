package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/northes/go-moonshot"
	"github.com/spf13/viper"
)

var (
	moonshotClient     *moonshot.Client
	moonshotClientOnce sync.Once
	lastKey            string
)

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("failed to load config: %v", err)
		return
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

	log.Fatal(app.Listen(fmt.Sprintf(":%s", viper.GetString(ConfigPath.Port))))
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

func HandleTranslation(c fiber.Ctx) (err error) {
	// 使用配置文件或环境变量初始化(一次)
	if moonshotClient == nil {
		key := viper.GetString(ConfigPath.Key)
		moonshotClientOnce.Do(func() {
			moonshotClient, err = moonshot.NewClient(key)
		})
		if err != nil {
			moonshotClientOnce = sync.Once{}
			return err
		}
	}

	// 使用Query初始化(每次)
	key := c.Query("key")
	// 简单对比新旧key是否相同
	if len(key) != 0 && key != lastKey {
		moonshotClient, err = moonshot.NewClient(key)
		if err != nil {
			return err
		}
		lastKey = key
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
				Content: viper.GetString(ConfigPath.Prompt),
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
