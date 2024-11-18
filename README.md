<p align="center"><img src="logo.png"></p>
<p align="center">A modern Lisp that compiles to JavaScript with first class interoperability.</p>

<h3 align="center">version 0.1.0 ALPHA</h3>

**Compiler goals**

- Interoperability with all Javascript environments (Node, Deno, Bun, browser etc)
- First class REPL (currently Deno) with realtime syntax highlighting and readline capabilities
- First class compiler error messages

**Language features**

- First class functions
- Pattern matching
- Destructuring
- Threading
- Macros

**Fennel inspired syntax**

```clojure
> (fn fib [n]
    (if (< n 2)
        n
        (+ (fib (- n 1))
           (fib (- n 2)))))

<fn fib>

> (fib 10)

55
```

**Destructuring**

```clojure
> (fn pair->sum [[x y]]
    "Returns the sum of two numbers in a vector."
    (+ x y))

<fn pair->sum>

> (pair->sum [1 4])

5
```

**Pattern matching**

```clojure
> (match [1 2 3]
    []      "empty list"
    [2]     "a single two"
    [_ _]   "two items"
    [1 _ 3] "one something three"
    :else   "unknown")

"one something three"
```

**Macros**

```clojure
> (macro for [[index start end next] body]
    `(do (var ,index ,start)
         (while (!= ,index ,end)
           (do ,body
               (,next ,index)))))

<macro for>

> (for [i 0 10 inc]
    (print (inc 1)))

1 2 3 4 5 6 7 8 9 10
```

**First class interop with any JS runtime**

```clojure
> (-> (. Deno (readTextFileSync "README.md"))
      (split-lines)
      (get 2))

"A modern Lisp that compiles to JavaScript with first class interoperability."
```

## Installation

Requires [Go 1.22](https://go.dev/dl/) and [Deno](https://deno.com/) as the compiler uses it as the default environment in which to run compiled code.

**Linux**

```bash
$ make install
```

**Others**

```bash
$ make build
$ mv rem $(SOME_PATH)
```
