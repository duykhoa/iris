//Package test -v ./... builds all tests
package test

// Contains tests for sessions(sessions package) & flash messages(context)

import (
	"testing"

	"github.com/kataras/iris"
)

func TestSessions(t *testing.T) {

	values := map[string]interface{}{
		"Name":   "iris",
		"Months": "4",
		"Secret": "dsads£2132215£%%Ssdsa",
	}

	api := iris.New()
	api.Config.Sessions.Cookie = "mycustomsessionid"

	writeValues := func(ctx *iris.Context) {
		sessValues := ctx.Session().GetAll()
		ctx.JSON(iris.StatusOK, sessValues)
	}

	if EnableSubdomainTests {
		api.Party(Subdomain+".").Get("/get", func(ctx *iris.Context) {
			writeValues(ctx)
		})
	}

	api.Post("set", func(ctx *iris.Context) {
		vals := make(map[string]interface{}, 0)
		if err := ctx.ReadJSON(&vals); err != nil {
			t.Fatalf("Cannot readjson. Trace %s", err.Error())
		}
		for k, v := range vals {
			ctx.Session().Set(k, v)
		}
	})

	api.Get("/get", func(ctx *iris.Context) {
		writeValues(ctx)
	})

	api.Get("/clear", func(ctx *iris.Context) {
		ctx.Session().Clear()
		writeValues(ctx)
	})

	api.Get("/destroy", func(ctx *iris.Context) {
		ctx.SessionDestroy()
		writeValues(ctx)
		// the cookie and all values should be empty
	})

	e := Tester(api, t)

	e.POST("/set").WithJSON(values).Expect().Status(iris.StatusOK).Cookies().NotEmpty()
	e.GET("/get").Expect().Status(iris.StatusOK).JSON().Object().Equal(values)
	if EnableSubdomainTests {
		e.Request("GET", SubdomainURL+"/get").Expect().Status(iris.StatusOK).JSON().Object().Equal(values)
	}

	// test destory which also clears first
	d := e.GET("/destroy").Expect().Status(iris.StatusOK)
	d.JSON().Object().Empty()
	d.Cookies().ContainsOnly(api.Config.Sessions.Cookie)
	// set and clear again
	e.POST("/set").WithJSON(values).Expect().Status(iris.StatusOK).Cookies().NotEmpty()
	e.GET("/clear").Expect().Status(iris.StatusOK).JSON().Object().Empty()
}

func FlashMessagesTest(t *testing.T) {
	api := iris.New()
	values := map[string]string{"name": "kataras", "package": "iris"}

	api.Put("/set", func(ctx *iris.Context) {
		for k, v := range values {
			ctx.SetFlash(k, v)
		}
	})

	//we don't get the flash so on the next request the flash messages should be available.
	api.Get("/get_no_getflash", func(ctx *iris.Context) {})

	api.Get("/get", func(ctx *iris.Context) {
		// one time one handler
		kv := make(map[string]string)
		for k := range values {
			kv[k], _ = ctx.GetFlash(k)
		}
		//second time on the same handler
		for k := range values {
			kv[k], _ = ctx.GetFlash(k)
		}

	}, func(ctx *iris.Context) {
		// third time on a next handler
		// test the if next handler has access to them(must) because flash are request lifetime now.
		kv := make(map[string]string)
		for k := range values {
			kv[k], _ = ctx.GetFlash(k)
		}
		// print them to the client for test the response also
		ctx.JSON(iris.StatusOK, kv)
	})

	e := Tester(api, t)
	e.PUT("/set").Expect().Status(iris.StatusOK).Cookies().NotEmpty()
	// just a request which does not use the flash message, so flash messages should be available on the next request
	e.GET("/get_no_getflash").Expect().Status(iris.StatusOK).Cookies().NotEmpty()
	e.GET("/get").Expect().Status(iris.StatusOK).JSON().Object().Equal(values)
	// second request ,the flash messages here should be not available and cookie has been removed
	// (the true is that the cookie is removed from the first GetFlash, but is available though the whole request saved on context's values for faster get, keep that secret!)*
	g := e.GET("/get").Expect().Status(iris.StatusOK)
	g.JSON().Object().Empty()
	g.Cookies().Empty()

}
