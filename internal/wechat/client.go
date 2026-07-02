package wechat

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"project20260702/internal/config"
)

const code2SessionURL = "https://api.weixin.qq.com/sns/jscode2session"

// Code2SessionResult 是微信 code2Session 接口返回的数据。
type Code2SessionResult struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// Code2Session 使用小程序 wx.login 拿到的 code 换取 openid。
//
// 本地开发时，如果没有配置真实 AppSecret，会返回 DevOpenID，方便先跑通登录链路。
func Code2Session(cfg config.WeChatConfig, code string) (Code2SessionResult, error) {
	if cfg.AppSecret == "" || cfg.AppID == "" || cfg.AppID == "touristappid" {
		return Code2SessionResult{
			OpenID: cfg.DevOpenID,
		}, nil
	}

	values := url.Values{}
	values.Set("appid", cfg.AppID)
	values.Set("secret", cfg.AppSecret)
	values.Set("js_code", code)
	values.Set("grant_type", "authorization_code")

	requestURL := fmt.Sprintf("%s?%s", code2SessionURL, values.Encode())

	resp, err := http.Get(requestURL)
	if err != nil {
		return Code2SessionResult{}, err
	}
	defer resp.Body.Close()

	var result Code2SessionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Code2SessionResult{}, err
	}

	if result.ErrCode != 0 {
		return Code2SessionResult{}, errors.New(result.ErrMsg)
	}

	if result.OpenID == "" {
		return Code2SessionResult{}, errors.New("wechat openid is empty")
	}

	return result, nil
}
