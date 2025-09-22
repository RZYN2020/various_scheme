((lambda (x) (+ x 1)) 10)
; expected: 11

(define add-one (lambda (x) (+ x 1)))
(add-one 99)
; expected: 100

(define make-adder
  (lambda (x)
    (lambda (y) (+ x y))))
(define add-five (make-adder 5))
(add-five 10)
; expected: 15

((lambda (x) ((lambda (y) (+ x y)) 20)) 10)
; expected: 30