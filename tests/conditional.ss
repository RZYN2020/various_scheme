(if #t 1 0)
; expected: 1

(if #f 1 0)
; expected: 0

(if (and #t #t) 1 0)
; expected: 1

(if (if #t #f #t) 1 0)
; expected: 0

(if (if #f #t #f) 1 0)
; expected: 0

(if (if #t #t #f) 1 0)
; expected: 1

(and #t #t)
; expected: #t

(and #t #f)
; expected: #f

(and #f #t)
; expected: #f

(and #f #f)
; expected: #f

(or #t #t)
; expected: #t

(or #t #f)
; expected: #t

(or #f #t)
; expected: #t

(or #f #f)
; expected: #f

(not #t)
; expected: #f

(not #f)
; expected: #t