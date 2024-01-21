
```haskell
data Iterator i o r =
    | Result r
    | Susp o ( i −> Iterator i o r )
type Yield = Iterator

yield :: o −> Yield i o i
yield v = Susp v return

instance Monad ( Yield i o ) where
	return = Result
	( Result v ) >>= f = f v
	( Susp v k ) >>= f = Susp v ( \ x −> ( k x )>>=f )

run :: Yield i o r −> Iterator i o r
run = id
```