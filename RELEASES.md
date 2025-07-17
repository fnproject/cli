
# Releases
## v 0.6.43

* Upgrade the Go version to 1.23.10 to mitigate the CVE-2025-22874
## v 0.6.41

* Add missing release note for v0.6.40

## v 0.6.40

* Small tweak to the OL8 cross build error message

## v 0.6.39

* Enabling support for go1.24
* Removing support for go1.20
 
## v 0.6.37

* Make fn-cli compatiable with Oracle Cloud Shell based on OL8

## v 0.6.35

* Enabling support for go1.23
* Removing support for go1.19

## v 0.6.34

* Enabling support for go1.20

## v 0.6.31

* Enabling support for node20
* Removing support for node16, node 14

## v 0.6.30
* 
* Fix security vulnerability issue
* Remove Ruby 2.7 support

## v 0.6.26

* Enabling support for node18, node16, go1.19, go1.18, ruby3.1.
* Removing support for node11, go1.15, python3.6, python3.7

## v 0.6.25

* Support for multiple shapes(architectures) functions images: 
  * x86 (default)
  * arm
  * multiarch (x86, arm)
  
* cli now is supported on `Arm Linux` as well.


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