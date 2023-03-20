
# Releases

## v 0.6.24

* Allowing `fn inspect functions` and `fn list functions` for PBF(Pre-Built Functions) function with empty image and digest field. By default, it was not supported. 
  
  Note: If you have functions created using Pre-Built Functions, then please upgrade to this version to have fn list and fn inspect  work properly


## v 0.6.7

* Support for following languages versions:
    * Node 14
    * Go 1.15
    * Ruby 2.7
  
    Check out `fn init --help` for available runtime environments.
* Docker runtime and build image stamping in func.yaml for a language runtime. 

## v 0.4.156

* Routes have now been removed from fn and replaced with functions and triggers.
* The migrate command will upgrade your func.yaml to include a trigger section in place of `path` field.
* `fn call` has been replaced with `fn invoke`.

Please see [Setting Functions Free Blog Post](https://medium.com/fnproject/setting-functions-free-15d063be72bf) and [Fn Project Tutorials](http://fnproject.io/tutorials/) for more information.