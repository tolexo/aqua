# AQUA
Golang Restful APIs in a cup, and ready to serve!


##Inspiration
- Apache and IIS WebServers
- Go-Rest API framework

## Design Goals

- Simplicity & Modularity
- Developer Productivity
- High Configurability
- Low learning curve
- Easy Versioning
- Pluggable Modules (using golang middleware)
- High Performance
- Preference for Json (over xml)

## Features

- Aqua uses service controllers to define related endpoints. This makes code modular and organized
- Using Aqua, you can define configurations at 4 levels:
  1. at server level, programmatically (these are inherited by all endpoints)
  2. at service controller level, declaratively using golang tags (these are inherited by all contained apis)
  3. at service controller level, programmatically
  4. at api or endpoint level, declaratively (these override inherited configurations)
- You can declaratively specify the version
  - Multiple versions are supported easily by
     - defining at service controller level (inherited by all internal endpoints)
     - overriding for each endpoint specifically
- Out-of-box support for common tasks like
  - Caching
  - Database binding (for CRUD operations) *|pending*
  - Working with Queues *|pending*
  - Stubbing
     - If there are code/project dependencies on your api service, you can simply write a stub (sample output) in an external file and publish this mock api quickly before writing actual business logic
- You can define modules (middleware) at a project level and then apply them to any service using Aqua's powerful configuration model

##Lets explore these features

### Q: How do I write a 'hello world' api?
First define a service controller in your project that supports a GET response (aqua.GetApi as its type). Note that the controller defined as a struct must anonymously include aqua.RestService. 

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi
}
```

Now implement a method corresponding to 'world' field after uppercasing the first letter. To start off, the method can return a string (more on this later).

```
func (me *HelloService) World() string {
	return "Hello World"
}
```

Now setup your main function to run the Aqua rest server

```
server := aqua.NewRestServer()
server.AddService(&HelloService{})
server.Run()
```

Now open your browser window, and hit http://localhost:8090/hello/world

---

### Q: But I don't need any magic; What about the unadulterated http requests and responses?

Sure, just change the function signature and you are good to go.

```
func (me *HelloService) World(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello There!")
}
```
---

###Q: I want to change the url from /hello/world to /hello/moon. Do I need to change the method names?

The service urls are derived from url tags. If none are specified then it defaults to the method name. So you can simply introduce the tag as follows. 

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi `url:"moon"`
}
```

---

### Q: What if I need to return both Hello World, and Hello There as different versions of the same GET api?

Simply add both the methods, but specify versions in field tags.

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi `version:"1.0" url:"moon"`
	worldNew aqua.GetApi `version:"1.1" url:"moon"`
}
func (me *HelloService) World() string {
	return "Hello World"
}
func (me *HelloService) WorldNew(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello There!")
}
```
Now you can hit:

http://localhost:8090/v1.0/hello/moon and

http://localhost:8090/v1.1/hello/moon to see the difference.

---

### Q: What options can I use to customize URLs for my apis? 

There are 3 out-of-box setting available, that help you customize URLs. 

- prefix
- root
- url

We have already seen how 'url' works. 

To change the root directory (*hello*), you can use the *root* tag at each service level, or more simply at the service controller level as demonstrated below:

```
type HelloService struct {
	aqua.RestService  `root:"this-is-the"`
	world aqua.GetApi `version:"1.0" url:"moon"` 
	worldNew aqua.GetApi `version:"1.1" url:"moon"`
}
```

With this change, your api endpoints are now working as:

*http://localhost:8090/v1.0/this-is-the/moon* and

*http://localhost:8090/v1.1/this-is-the/moon*

You can also use the 'prefix' field. This part comes in before version information in the final constructed endpoint url

```
type HelloService struct {
	aqua.RestService  `root:"this-is-the" prefix:"sunshine"`
	world aqua.GetApi `version:"1.0" url:"moon"` 
	worldNew aqua.GetApi `version:"1.1" url:"moon"`
}
```
So with this prefix now set, our end points would become:

*http://localhost:8090/sunshine/v1.0/this-is-the/moon*

*http://localhost:8090/sunshine/v1.1/this-is-the/moon*

Also note that, all there of these properties (url, root and prefix) can contain any number of slashes. So if you change the url to:

```
type HelloService struct {
	aqua.RestService  `root:"this-is-the" prefix:"sunshine"`
	world aqua.GetApi `version:"1.0" url:"/good/old/moon"` 
}
```

Then you get the final url as:

http://localhost:8090/sushine/v1.0/this-is-the/good/old/moon. 

### Q: Does Aqua use any mux?

Yes, Gorilla mux is used internally. So to define url parameters, we'll need to follow Gorilla mux conventions. We'll get to those in a moment

### Q: How can I check if the server is up and running?

By default an "aqua" route is setup:

 - */aqua/ping* returns "pong" if the server is running
 - */aqua/status* returns version, go runtime memory information
 - */aqua/time* returns current server time

### Q: What is the default port that Aqua runs on?

It's 8090. You can change it though as follows:

```
server := aqua.NewRestServer()
server.AddService(&HelloService{})
server.Port = 5432;
server.Run()
```

### Q: When I use api versioning, can I use HTTP headers to pass the version info?

```
type CatalogService struct {
	aqua.RestService  `root:"catalog" prefix:"mycompany"`
	getProduct aqua.GetApi `version:"1.0" url:"product"` 
}
```

If you setup a catalog service as shown above then out of box you can use version capability as shown below

1. GET call to http://localhost:8090/mycompany/v1.0/catalog/product
2. GET call to http://localhost:8090/mycompany/catalog/product
  - pass a request header "Accept": "application/vnd.api+json;version=1.0"
  -  *-or-*
  - pass a request header "Accept": "application/vnd.api-v1.0+json"

Note: If you want to customize the media type, you can do so. 

```
type CatalogService struct {
	aqua.RestService  `root:"catalog" prefix:"mycompany"`
	getProduct aqua.GetApi `vendor:"vnd.myorg.myfunc.api" version:"1.0" url:"product"` 
}
```
Basis this, the required Accept header will be need to changed to following:

- "Accept" header : "application/__vnd.myorg.myfunc.api__+json;version=1.0"
-  *-or-*
- "Accept" header : "application/__vnd.myorg.myfunc.api__-v1.0+json"


### Q: How can I access query strings?

Its simple, you add an input variable to your implementation method of type aqua.Jar. This variable gives you access to the Request object, and also has some helper method as shown below:

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi
}

func (me *HelloService) World(j aqua.Jar) string {
	j.LoadVars()
	return "Hello " + j.QueryVars["country"]
}
```

Now, just hit the url: http://localhost:8090/hello/world?country=Singapore

### Q: How do I pass dynamic parameters to apis?

You start by defining the url with the appropriate dynamic variable as per the guidelines of Gorilla mux. 

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi `url:"/country/{c}"`
}
```
Then you just read this value in the associated method. Note: Aqua currently supports passing int and string parameters. 

```
func (me *HelloService) World(c string) string {
	return "Hello " + c
}
```

Now, you can hit http://localhost:8090/hello/country/Brazil

In case you are reading an integer value, then you can define strict logic in url to only match numbers using a regular expression:

```
type HelloService struct {
	aqua.RestService
	world aqua.GetApi `url:"/country/{c}"`
	capital aqua.GetApi `url:/capital/{cap:[0-9]+}`
}
```


### Q: Can you explain how the configuration model works? Will I need to define attributes at each endpoint level?

Aqua has a powerful configuration model that works at 4 levels:

1. Server (programmatically)
2. Service controller (declaratively)
3. Service controller (programmatically)
4. Endpoint (declaratively)

Lets look at each of them in detail

##### 1. Server (programmatically)

If you define any configuration at the server level, then it is __inherited__ by all the Service controllers and all the contained services automatically.

```
server := aqua.NewRestServer()

// Note:
server.Prefix = "myapis"
// Prefix value is inherited by everything on this server!!

server.AddService(&HelloService{})
server.AddService(&HolaService{})
server.Run()
```
##### 2. Service controller (declaratively)

We added two service controllers to the server above - HelloService and HolaService. Let's assume that all the contained services need to begin with words 'Hello' and 'Hola' respectively. 

To achive this, we specify the 'root' variable at the top level by defining it agains the RestServer.

```
type HelloService struct {
	aqua.RestService `root:"Hello"`
	service1 aqua.GetApi
	service2 aqua.GetApi
	..
	serviceN aqua.GetApi
}
```

This ensures that all services in this now __inherit__ the root value of "Hello"

##### 4. Endpoint (declaratively)

Last but not the least, you can specify a value at a service endpoint. You can do so by configuring at the api level as shown below. Note that these values will override the inherited values.

```
type HelloService struct {
	aqua.RestService `root:"Hello"`
	service1 aqua.GetApi `root:"Hiya"` //Hiya overrides Hello
	service2 aqua.GetApi
	..
	serviceN aqua.GetApi
}
```

###Q: What all configurations are available in Aqua?



### Q: How do I enable caching for my apis?

### Q: What are 'modules' and how can I use them?

### Q: Can I create mock apis?

### Q: What are different return types supported by default?

