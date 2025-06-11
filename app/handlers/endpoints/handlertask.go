package endpoints

import (
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/yetiz-org/gone/channel"
	"github.com/yetiz-org/gone/http"
	"github.com/yetiz-org/gone/http/httpheadername"
	"github.com/yetiz-org/gone/http/httpstatus"
	buf "github.com/yetiz-org/goth-bytebuf"
	"github.com/yetiz-org/goth-kkdatastore"
	"github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-kktemplate"
	"github.com/yetiz-org/goth-kktranslation"
	"github.com/yetiz-org/goth-scaffold/app/conf"
	"github.com/yetiz-org/goth-scaffold/app/constant/page"
	"github.com/yetiz-org/goth-scaffold/app/constant/param"
	"github.com/yetiz-org/goth-scaffold/app/constant/query"
)

type HandlerTask struct {
	http.DefaultHTTPHandlerTask
}

func (h *HandlerTask) PreCheck(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *HandlerTask) Index(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return http.NotImplemented
}

func (h *HandlerTask) Get(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Post(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Put(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Delete(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Options(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *HandlerTask) Patch(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Trace(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Connect(ctx channel.HandlerContext, req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *HandlerTask) Before(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *HandlerTask) After(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *HandlerTask) ErrorCaught(req *http.Request, resp *http.Response, params map[string]interface{}, err http.ErrorResponse) error {
	resp.ResponseError(err)
	return nil
}

func (h *HandlerTask) ResponseError(er http.ErrorResponse, resp *http.Response) {
	resp.ResponseError(er)
}

func (h *HandlerTask) Redirect(redirectUrl string, resp *http.Response) {
	resp.Redirect(redirectUrl)
}

func (h *HandlerTask) RenderHtml(templateName string, config *RenderConfig, resp *http.Response) {
	if tmpl, err := kktemplate.LoadFrameHtml(templateName, h.Lang(resp.Request())); err == nil {
		buffer := buf.EmptyByteBuf()
		renderVars := h._RenderVars(templateName, config, resp)
		if e := tmpl.ExecuteTemplate(buffer, "main", renderVars); e != nil {
			kklogger.ErrorJ("HandlerTask.RenderHtml", fmt.Sprintf("ExecuteTemplate Fail, Err: %s", e.Error()))
		}

		resp.SetHeader(httpheadername.ContentType, mime.TypeByExtension(".html"))
		resp.SetBody(buffer)
	} else {
		kklogger.ErrorJ("HandlerTask.RenderHtml", err.Error())
	}
}

func (h *HandlerTask) ReaderDB() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Reader().DB()
}

func (h *HandlerTask) WriterDB() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Writer().DB()
}

func (h *HandlerTask) RedisWDB() redis.Conn {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Master().Conn()
}

func (h *HandlerTask) RedisRDB() redis.Conn {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Slave().Conn()
}

func (h *HandlerTask) T(message string, lang string) string {
	return kktranslation.GetLangFile(lang).T(message)
}

func (h *HandlerTask) Lang(req *http.Request) string {
	lang := strings.ToLower(req.FormValue(query.Lang))

	if lang == "" {
		if sessionLang := req.Session().GetString(param.Lang); sessionLang != "" {
			return sessionLang
		}
	} else {
		if kktranslation.GetLangFile(lang) != nil {
			req.Session().PutString(param.Lang, lang)
			return lang
		}
	}

	for _, qv := range req.AcceptLanguage() {
		lang = strings.ToLower(qv.Value.String())
		if kktranslation.GetLangFile(lang) != nil {
			req.Session().PutString(param.Lang, lang)
			return lang
		}
	}

	lang = strings.ToLower(conf.Config().Lang.Default)
	req.Session().PutString(param.Lang, lang)

	return lang
}

type RenderConfig struct {
	PageTitle        string
	JavascriptHeader string
	JavascriptFooter string
	PageRenderData   map[string]interface{}
}

func (h *HandlerTask) _RenderVars(pageID string, config *RenderConfig, resp *http.Response) map[string]interface{} {
	if config == nil {
		config = &RenderConfig{}
	}

	lang := h.Lang(resp.Request())
	redirect := resp.Request().FormValue(query.Redirect)
	renderVars := map[string]interface{}{}
	renderVars["Time_Now"] = time.Now()
	renderVars["Page_ID"] = pageID
	renderVars["Page_Title"] = h.T(config.PageTitle, h.Lang(resp.Request()))
	renderVars["RequestPath"] = resp.Request().Url().Path
	renderVars["RequestUri"] = resp.Request().RequestURI()
	renderVars["Lang"] = lang
	if langFile := kktranslation.GetLangFile(lang); langFile != nil {
		renderVars["LangName"] = langFile.Name
	} else {
		renderVars["LangName"] = kktranslation.GetLangFile(conf.Config().Lang.Default).Name
	}

	renderVars["LangFiles"] = kktranslation.LangFiles()
	remoteIp, remotePort := resp.Request().RemoteAddr()
	renderVars["RemoteIP"] = remoteIp
	renderVars["RemotePort"] = remotePort
	renderVars["Redirect"] = redirect
	renderVars["Javascript_Header"] = config.JavascriptHeader
	renderVars["Javascript_Footer"] = config.JavascriptFooter

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
		m := map[string]interface{}{}
		session.GetStruct(page.RenderData, &m)
		if len(m) > 0 {
			for k, v := range m {
				if s, ok := v.(string); ok {
					m[k] = h.T(s, h.Lang(resp.Request()))
				} else {
					m[k] = v
				}
			}

			renderVars["SessionData"] = m
		}

		session.Delete(page.RenderData)
	}

	return renderVars
}

func (h *HandlerTask) SessionRenderData(resp *http.Response, key string, value interface{}) {
	if resp != nil {
		session := resp.Request().Session()
		m := map[string]interface{}{}
		session.GetStruct(page.RenderData, &m)
		m[key] = value
		session.PutStruct(page.RenderData, m)
	}
}
