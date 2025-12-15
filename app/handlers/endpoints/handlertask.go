package endpoints

import (
	"encoding/json"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/ghttp"
	"github.com/yetiz-org/gone/ghttp/httpheadername"
	"github.com/yetiz-org/gone/ghttp/httpstatus"
	buf "github.com/yetiz-org/goth-bytebuf"
	kklogger "github.com/yetiz-org/goth-kklogger"
	kktemplate "github.com/yetiz-org/goth-kktemplate"
	kktranslation "github.com/yetiz-org/goth-kktranslation"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/constant/page"
	"github.com/yetiz-org/goth-scaffold/app/constant/query"
	"github.com/yetiz-org/goth-util/hash"
)

type HandlerTask struct {
	ghttp.DefaultHTTPHandlerTask
}

func (h *HandlerTask) Super() ghttp.HttpHandlerTask {
	return &h.DefaultHTTPHandlerTask
}

func (h *HandlerTask) PreCheck(req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	return nil
}

func (h *HandlerTask) Index(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	return ghttp.NotImplemented
}

func (h *HandlerTask) Get(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Post(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Put(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Delete(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Options(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	return nil
}

func (h *HandlerTask) Patch(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Trace(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Connect(ctx channel.HandlerContext, req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Before(req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	return nil
}

func (h *HandlerTask) After(req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}) ghttp.ErrorResponse {
	return nil
}

func (h *HandlerTask) ErrorCaught(req *ghttp.Request, resp *ghttp.Response, params map[string]interface{}, err ghttp.ErrorResponse) error {
	m := map[string]any{}
	if jErr := json.Unmarshal([]byte(err.Message()), &m); jErr == nil {
		err.ErrorData()[".message"] = m
	}

	if resp.StatusCode() == 0 {
		resp.ResponseError(err)
	}

	return nil
}

func (h *HandlerTask) GetNode(params map[string]any) ghttp.RouteNode {
	if rtn := params["[gone-http]node"]; rtn != nil {
		return rtn.(ghttp.RouteNode)
	}

	return nil
}

func (h *HandlerTask) GetNodeId(params map[string]any) string {
	return h.GetID(h.GetNodeName(params), params)
}

func (h *HandlerTask) Host(req *ghttp.Request) string {
	if host := req.Host(); host != "" {
		return host
	} else {
		return conf.Config().App.DomainName.String()
	}
}

func (h *HandlerTask) ResponseError(er ghttp.ErrorResponse, resp *ghttp.Response) {
	resp.ResponseError(er)
}

func (h *HandlerTask) GenerateCSRFToken(data []byte, expiresIn time.Duration) (csrfToken string) {
	csrfToken = hash.CryptoTimeHash(data, time.Now().Add(expiresIn).Unix(), []byte("csrf"))
	return
}

func (h *HandlerTask) ValidateCSRFToken(csrfToken string) (data []byte, valid bool) {
	if !hash.ValidateTimeHash(csrfToken) || hash.TimestampOfTimeHash(csrfToken) < time.Now().Unix() {
		return nil, false
	}

	data = hash.DataOfCryptoTimeHash(csrfToken, []byte("csrf"))
	return data, true
}

func (h *HandlerTask) Redirect(redirectUrl string, resp *ghttp.Response) {
	resp.Redirect(redirectUrl)
}

func (h *HandlerTask) RenderHtml(templateName string, config *RenderConfig, resp *ghttp.Response) {
	if tmpl, err := kktemplate.LoadFrameHtml(templateName, h.Lang(resp.Request())); err == nil {
		buffer := buf.EmptyByteBuf()
		renderVars := h._RenderVars(templateName, config, resp)
		if e := tmpl.ExecuteTemplate(buffer, "main", renderVars); e != nil {
			kklogger.ErrorJ("endpoints:HandlerTask.RenderHtml#template!execute_fail", fmt.Sprintf("ExecuteTemplate Fail, Err: %s", e.Error()))
		}

		resp.SetHeader(httpheadername.ContentType, mime.TypeByExtension(".html"))
		resp.SetBody(buffer)
	} else {
		kklogger.ErrorJ("endpoints:HandlerTask.RenderHtml#template!load_fail", err.Error())
	}
}

func (h *HandlerTask) T(message string, lang string) string {
	return kktranslation.GetLangFile(lang).T(message)
}

func (h *HandlerTask) Lang(req *ghttp.Request) string {
	lang := strings.ToLower(req.FormValue(query.Lang))

	if lang == "" {
		if sessionLang := req.Session().GetString(query.Lang); sessionLang != "" {
			return sessionLang
		}
	} else {
		if kktranslation.GetLangFile(lang) != nil {
			req.Session().PutString(query.Lang, lang)
			return lang
		}
	}

	for _, qv := range req.AcceptLanguage() {
		lang = strings.ToLower(qv.Value.String())
		if kktranslation.GetLangFile(lang) != nil {
			req.Session().PutString(query.Lang, lang)
			return lang
		}
	}

	lang = strings.ToLower(conf.Config().Lang.Default)
	req.Session().PutString(query.Lang, lang)

	return lang
}

type RenderConfig struct {
	PageTitle        string
	JavascriptHeader string
	JavascriptFooter string
	PageRenderData   map[string]interface{}
}

func (h *HandlerTask) SetPageErrorMessage(message string, resp *ghttp.Response) {
	resp.Request().Session().PutString(page.ErrorMessage, message)
}

func (h *HandlerTask) SetPageSuccessMessage(message string, resp *ghttp.Response) {
	resp.Request().Session().PutString(page.SuccessMessage, message)
}

func (h *HandlerTask) _RenderVars(pageID string, config *RenderConfig, resp *ghttp.Response) map[string]interface{} {
	if config == nil {
		config = &RenderConfig{}
	}

	lang := h.Lang(resp.Request())
	redirect := resp.Request().FormValue(query.Redirect)
	renderVars := map[string]interface{}{}
	renderVars["TimeNow"] = time.Now()
	renderVars["PageID"] = pageID
	renderVars["PageTitle"] = h.T(config.PageTitle, h.Lang(resp.Request()))
	renderVars["RequestPath"] = resp.Request().Url().Path
	renderVars["RequestUri"] = resp.Request().RequestURI()
	renderVars["Lang"] = lang
	if langFile := kktranslation.GetLangFile(lang); langFile != nil {
		renderVars["LangName"] = langFile.Name
	} else {
		renderVars["LangName"] = kktranslation.GetLangFile(conf.Config().Lang.Default).Name
	}

	renderVars["GTMId"] = conf.Config().Credentials.GTMId
	renderVars["LangFiles"] = kktranslation.LangFiles()
	remoteIp, remotePort := resp.Request().RemoteAddr()
	renderVars["RemoteIP"] = remoteIp
	renderVars["RemotePort"] = remotePort
	renderVars["Redirect"] = redirect
	renderVars["JavascriptHeader"] = config.JavascriptHeader
	renderVars["JavascriptFooter"] = config.JavascriptFooter
	renderVars["Environment"] = conf.Config().App.Environment.Upper()
	renderVars["IsDebug"] = conf.IsDebug()
	renderVars["IsProduction"] = conf.IsProduction()
	renderVars["IsStaging"] = conf.IsStaging()
	renderVars["IsLocal"] = conf.IsLocal()

	if config.PageRenderData != nil {
		m := map[string]interface{}{}
		for k, v := range config.PageRenderData {
			if s, ok := v.(string); ok {
				m[k] = h.T(s, h.Lang(resp.Request()))
			} else {
				m[k] = v
			}
		}

		renderVars["PageData"] = m
	}

	if session := resp.Request().Session(); session != nil {
		renderVars["Session"] = session.Data()
	}

	if resp.Request().Session().GetString(page.ErrorMessage) != "" {
		renderVars["ErrorMessage"] = resp.Request().Session().GetString(page.ErrorMessage)
		resp.Request().Session().Delete(page.ErrorMessage)
	}

	if resp.Request().Session().GetString(page.SuccessMessage) != "" {
		renderVars["SuccessMessage"] = resp.Request().Session().GetString(page.SuccessMessage)
		resp.Request().Session().Delete(page.SuccessMessage)
	}

	return renderVars
}
