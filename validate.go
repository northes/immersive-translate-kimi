package main

import (
	"github.com/go-playground/validator/v10"
)

type structValidator struct {
	validate *validator.Validate
}

func (s *structValidator) Engine() any {
	return s.validate
}

func (s *structValidator) ValidateStruct(out any) error {
	return s.validate.Struct(out)
}
