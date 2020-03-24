package context

type CtxKey string

var AppContextKey = CtxKey("app")
var RequestContextKey = CtxKey("request")
var ClientContextKey = CtxKey("client")
