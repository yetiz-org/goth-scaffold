package endpoints

import (
	"bytes"
	"fmt"
	"mime"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/kklab-com/gone-httpheadername"
	"github.com/kklab-com/gone-httpstatus"
	"github.com/kklab-com/gone/http"
	"github.com/kklab-com/goth-kkdatastore"
	"github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kktemplate"
	"github.com/kklab-com/goth-kktranslation"
	"github.com/kklab-com/goth-scaffold/app/conf"
	"github.com/kklab-com/goth-scaffold/app/constant/page"
	"github.com/kklab-com/goth-scaffold/app/constant/param"
	"github.com/kklab-com/goth-scaffold/app/constant/query"
)

type KKHandlerTask struct {
	http.DefaultHandlerTask
}

func (h *KKHandlerTask) PreCheck(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *KKHandlerTask) Index(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return http.NotImplemented
}

func (h *KKHandlerTask) Get(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *KKHandlerTask) Post(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *KKHandlerTask) Put(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *KKHandlerTask) Delete(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *KKHandlerTask) Options(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *KKHandlerTask) Patch(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *KKHandlerTask) Trace(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *KKHandlerTask) Connect(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	resp.SetStatusCode(httpstatus.MethodNotAllowed)
	return nil
}

func (h *KKHandlerTask) Before(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *KKHandlerTask) After(req *http.Request, resp *http.Response, params map[string]interface{}) http.ErrorResponse {
	return nil
}

func (h *KKHandlerTask) ErrorCaught(req *http.Request, resp *http.Response, params map[string]interface{}, err http.ErrorResponse) error {
	resp.ResponseError(err)
	return nil
}

func (h *KKHandlerTask) ResponseError(er http.ErrorResponse, resp *http.Response) {
	resp.
		ResponseError(er)
}

func (h *KKHandlerTask) Redirect(redirectUrl string, resp *http.Response) {
	resp.Redirect(redirectUrl)
}

func (h *KKHandlerTask) RenderHtml(templateName string, config *RenderConfig, resp *http.Response) {
	if tmpl := kktemplate.LoadFrameHtml(templateName, h.Lang(resp.Request())); tmpl != nil {
		buffer := bytes.NewBuffer([]byte{})
		renderVars := h._RenderVars(templateName, config, resp)
		if e := tmpl.ExecuteTemplate(buffer, "main", renderVars); e != nil {
			kklogger.ErrorJ("HandlerTaskTemplate", fmt.Sprintf("ExecuteTemplate Fail, Err: %s", e.Error()))
		}

		resp.SetHeader(httpheadername.ContentType, mime.TypeByExtension(".html"))
		resp.SetBody(buffer)
	}
}

func (h *KKHandlerTask) ReaderDB() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Reader().DB()
}

func (h *KKHandlerTask) WriterDB() *gorm.DB {
	return datastore.KKDB(conf.Config().DataStore.DatabaseName).Writer().DB()
}

func (h *KKHandlerTask) RedisWDB() redis.Conn {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Master().Conn()
}

func (h *KKHandlerTask) RedisRDB() redis.Conn {
	return datastore.KKREDIS(conf.Config().DataStore.RedisName).Slave().Conn()
}

func (h *KKHandlerTask) T(message string, lang string) string {
	return kktranslation.GetLangFile(lang).T(message)
}

func (h *KKHandlerTask) Lang(req *http.Request) string {
	lang := strings.ToLower(req.Form.Get(query.Lang))

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
	DelayRedirectUri string
	DelaySec         int
}

func (h *KKHandlerTask) _RenderVars(pageID string, config *RenderConfig, resp *http.Response) map[string]interface{} {
	if config == nil {
		config = &RenderConfig{}
	}

	lang := h.Lang(resp.Request())
	redirect := resp.Request().Form.Get(query.Redirect)
	renderVars := map[string]interface{}{}
	renderVars["Now"] = time.Now()
	renderVars["PageID"] = pageID
	renderVars["PageTitle"] = h.T(config.PageTitle, h.Lang(resp.Request()))
	renderVars["RequestPath"] = resp.Request().URL.Path
	renderVars["RequestUri"] = resp.Request().RequestURI
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
	session := resp.Request().Session()
	renderVars["DelayRedirectUri"] = config.DelayRedirectUri
	renderVars["DelaySec"] = config.DelaySec

	m := map[string]interface{}{}
	session.GetStruct(page.RenderData, &m)
	if len(m) > 0 {
		for k, v := range m {
			if s, ok := v.(string); ok {
				m[k] = h.T(s, h.Lang(resp.Request()))
			}
		}

		renderVars["Data"] = m
	}

	session.Delete(page.RenderData)
	return renderVars
}

func (h *KKHandlerTask) PageRenderData(key string, value interface{}, resp *http.Response) {
	if resp != nil {
		session := resp.Request().Session()
		m := map[string]interface{}{}
		session.GetStruct(page.RenderData, &m)
		m[key] = value
		session.PutStruct(page.RenderData, m)
	}
}

func (h *KKHandlerTask) Javascript(javascriptCode string, resp *http.Response) {
	if resp != nil {
		resp.Request().Session().PutString(page.Javascript, javascriptCode)
	}
}
