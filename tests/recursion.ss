(define fib
  (lambda (n)
    (if (<= n 1)
        n
        (+ (fib (- n 1)) (fib (- n 2))))))

(fib 0)
; expected: 0

(fib 1)
; expected: 1

(fib 2)
; expected: 1

(fib 5)
; expected: 5

(fib 10)
; expected: 55

(fib 20)
; expected: 6765

(define factorial
  (lambda (n)
    (if (= n 0)
        1
        (* n (factorial (- n 1))))))

(factorial 0)
; expected: 1

(factorial 1)
; expected: 1

(factorial 5)
; expected: 120

(factorial 10)
; expected: 3628800

(define sum
  (lambda (n)
    (if (= n 0)
        0
        (+ n (sum (- n 1))))))

(sum 10)
; expected: 55

(sum 100)
; expected: 5050