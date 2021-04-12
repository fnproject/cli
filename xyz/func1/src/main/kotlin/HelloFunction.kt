
package com.fn.example

fun hello(input: String): String {
    println("Inside Kotlin Hello World function")
    return when {
        input.isEmpty() -> ("Hello, world!")
            else -> ("Hello, ${input}")
    }
}