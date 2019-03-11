Assignment 
==========

* You are to design a key-value store application, complete with tests, local deployment environment and documentation 

* while we have a strong preference for solutions provided in Python, we are accepting solutions in other languages suited for development of backend services like Node.JS, Go, Java or Scala - if you use one of these please mention in the README why you chose to do so 

* The application can be interfaced with via HTTP, its endpoints providing the following functionalities: 

    * get a value (GET /keys/{id}) 
    * get all values (GET /keys) 
    * set a value (PUT /keys) 
    * check if a value exists (HEAD /keys/{id}) 
    * delete a value (DELETE /keys/{id}) 
    * delete all values (DELETE /keys) 
    * set an expiry time when adding a value (PUT /keys?expire_in=60) 
    * support wildcard keys when getting all values (GET /keys?filter=wo$d) (the $ symbol should expand to match any number of characters, e.g: wod, word, world etc.) 

* The application will log any output to stdout 

* The application will use HTTP status codes to communicate success or failure of an operation 

* The data stored will be persisted so that restarts of the application don’t clear it 

* To integrate into our monitoring infrastructure, it is expected that the solution integrates with our Prometheus monitoring solution. Please instrument your application accordingly. 

* The solution should be provided as a single archive, containing a git repository and a Docker Compose file to start the application 

* send your solution in an archive, explicitly please do NOT upload it to GitHUB or other platform as a public repository 

Criteria 
========

* the solution works as expected, the tests you provide pass 

* the code style makes it easily maintainable, extendable and is according to the general best practices of the chosen platform 

* appropriate use of logging and metrics 

* documentation 

    * of the solution (is present and good enough to give an overview) 

    * of the code (is present where needed and not where obvious) 
    
    * includes any known limitations or restrictions of the solution 

    * mentions the platform tested on 

Remarks 
=======

* don’t spend time optimizing, try to focus on solving the problem 
* if you know that optimizations could be done, feel free to list them in a README 
* provide a README file (markdown or plain text) 
