# aqua
Golang Restful APIs in a cup, and ready to serve!



##Goals

1. Simplicity & Modularity
2. Low learning curve (developer usability)
3. Hierarchical yet granular control (high configurability)
   - configurations can be defined at multiple levels 
     - server level, programmatically (inherited by everything)
     - service level, declaratively (inherited by all api's of that service)
     - service level, programmatically
     - api or end point level, declaratively (applies to that particular service only)
4. Easy versioning
	- declaratively specify the version of api
	- support multiple versions by 
	  - adding new services and defining a service level version which then gets inherited to all apis of that service controller
	  - specify multiple functions within a rest service configured to work for different versions
5. Preference for Json (over Xml)
6. Developers (can & should) focus on high/object-level data structures
7. Re-usability (to call services directly or by bundling within an app)
8. Out-of-box DB binding support
9. Easy caching
10. High performance
