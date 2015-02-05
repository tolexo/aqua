# AQUA
Golang Restful APIs in a cup, and ready to serve!


##Inspiration
- Apache and IIS WebServers
- Go-Rest API framework

##Project Goals

-  Simplicity & Modularity
   -  Aqua uses service controllers to define related endpoints in a module
- Low learning curve (developer usability)
- High Configurability
   - configurations can be defined at 4 levels 
     - at server level, programmatically (inherited by everything)
     - at service level, declaratively (inherited by all api's in that service)
     - service level, programmatically)
     - api or end point level, declaratively (applies to that particular service only)
- Easy versioning
	- declaratively specify the version of an api
	- support multiple versions by 
	  - defining it at the service controller level (inherited by all internal endpoints)
	  - configuring different end points within a service controller to have different versions
- Preference for json (over xml)
- Developers (can & should) focus on high/object level data structures
- Out-of-box support for commong functionalities including
   - Binding to a DB
   - Working with Queues
- Easy caching
- High performance


### Q: How do I write a 'hello world' api?
First define a service controller in your project that supports a GET response (aqua.GetApi)

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi
}
```

Now implement a method corresponding to 'world' field after uppercasing the first letter. To start off, the method can return a string

```
func (me *HelloService) World() string {
	return "Hello World"
}
```

Now setup your main function to run the rest server

```
server := aqua.NewRestServer()
server.AddService(&HelloService{})
server.Run()
```

Now open your browser window, and hit *http://localhost:8080/hello/world*

---

### Q: But I don't need any magic; What about the unadulterated http requests and responses?

Sure, just change the function signature:

```
func (me *HelloService) World(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello There!")
}
```
---
### Q: What if need to return both Hello World, and Hello There as different versions of the same GET api?

Simply add both the methods, but specify versions in field tags.

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi `version:"1.0"`
	worldNew aqua.GetApi `version:"1.1"`
}
func (me *HelloService) World() string {
	return "Hello World"
}
func (me *HelloService) WorldNew(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello There!")
}
```
Now you can hit:

*http://localhost:8080/v1.0/hello/world* and

*http://localhost:8080/v1.1/hello/world* to see the difference.

---

### Q: How do I specify or customize URLs for my apis? 

There are 3 out-of-box setting available in field tags, that help you customize URLs. 

- prefix
- root
- url

Lets see how each of these work. To change the url after the main directory, you simply use the url tag.

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi `url:"aqua-world"` 
	worldNew aqua.GetApi `url:"water-world"`
}
```
After this change the *hello/world* urls won't work any more. Instead you'd need to use the following ones:

*http://localhost:8080/hello/aqua-world* and

*http://localhost:8080/hello/water-world*

To change the root directory (*hello*), you can use the *root* tag at each service level, or more simply at the service controller level as demonstrated below:

```
type HelloService struct {
	aqua.RestService  `root:"this-is"`
	world aqua.GetApi `url:"aqua-world"` 
	worldNew aqua.GetApi `url:"water-world"`
}
```

With this change, your api endpoints are now working at:

*http://localhost:8080/this-is/aqua-world* and

*http://localhost:8080/this-is/water-world*

### Q: Does Aqua use any mux?

Yes, Gorilla mux. And so to define url parameters, we'll need to follow Gorilla mux conventions. We'll get to those in a moment

