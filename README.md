# Remlisp

A modern Lisp that compiles to JavaScript with first class interop.


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

"A modern Lisp that compiles to JavaScript with first class interop."
```
