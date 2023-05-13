package model

import (
	"encoding/base32"
	"encoding/hex"
	"strings"

	"github.com/xlzd/gotp"
)

type App struct {
	Id                string `json:"_id"`
	Name              string `json:"name"`
	SerialId          int    `json:"serial_id"`
	Version           int    `json:"version"`
	AssetsGroup       string `json:"assets_group"`
	BackgroundColor   any    `json:"background_color"`
	CircleBackground  any    `json:"circle_background"`
	CircleColor       any    `json:"circle_color"`
	CustomAssets      bool   `json:"custom_assets"`
	GeneratingAssets  bool   `json:"generating_assets"`
	LabelsColor       any    `json:"labels_color"`
	LabelsShadowColor any    `json:"labels_shadow_color"`
	TimerColor        string `json:"timer_color"`
	TokenColor        any    `json:"token_color"`
	AuthyId           int    `json:"authy_id"`
	SecretSeed        string `json:"secret_seed"`
	Digits            int    `json:"digits"`
	MemberSince       int    `json:"member_since"`
	TransactionalOtp  bool   `json:"transactional_otp"`
}

func (a App) TOTP() (*gotp.TOTP, error) {
	dec, err := hex.DecodeString(a.SecretSeed)
	if err != nil {
		return nil, err
	}

	return gotp.NewTOTP(strings.TrimRight(base32.StdEncoding.EncodeToString(dec), "="), 7, 20, nil), nil
}
