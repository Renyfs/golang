package One

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type HandelFunc func(*Context)

type Engine struct {
	*RouterGroup
	router        *router
	groups        []*RouterGroup
	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
}
type RouterGroup struct {
	prefix      string
	middlewares []HandelFunc // support middleware
	parent      *RouterGroup // support nesting
	engine      *Engine      // all groups share a Engine instance
}

// New 初始化 Engine 实例
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}
func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

// -------------------------Engine--------------------------------
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) addRoute(method string, pattern string, handler HandelFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET GET方法定义
func (engine *Engine) GET(path string, handle HandelFunc) {
	engine.addRoute("GET", path, handle)
}

// POST POST方法定义
func (engine *Engine) POST(path string, handle HandelFunc) {
	engine.addRoute("POST", path, handle)
}

// PUT PUT方法定义
func (engine *Engine) PUT(path string, handle HandelFunc) {
	engine.addRoute("PUT", path, handle)
}

// DELETE DELETE方法定义
func (engine *Engine) DELETE(path string, handle HandelFunc) {
	engine.addRoute("DELETE", path, handle)
}

// Run 启动服务
func (engine *Engine) Run(addr string) (err error) {
	fmt.Printf("server listen %s port ...\n", addr)
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandelFunc
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.engine = engine
	c.handlers = middlewares
	engine.router.handle(c)
}

// Group ------------------------- Group --------------------------------
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandelFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}
func (group *RouterGroup) GET(pattern string, handler HandelFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandelFunc) {
	group.addRoute("POST", pattern, handler)
}
func (group *RouterGroup) PUT(pattern string, handler HandelFunc) {
	group.addRoute("PUT", pattern, handler)
}

func (group *RouterGroup) DELETE(pattern string, handler HandelFunc) {
	group.addRoute("DELETE", pattern, handler)
}
func (group *RouterGroup) Use(middlewares ...HandelFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// --------------------------Static--------------------------------------
// createStaticHandler create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandelFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check if file exists and/or if we have permission to access it
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static serve static files
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register GET handlers
	group.GET(urlPattern, handler)
}
