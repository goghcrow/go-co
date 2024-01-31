
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


```text
https://groups.google.com/g/elm-discuss/c/rAfKkv2w1GU

                  (a -> b) -> a -> b                Names: apply, <|, $
Functor f      => (a -> b) -> f a -> f b            Names: map, fmap, <$>, <~
Applicative f  => f (a -> b) -> f a -> f b          Names: ap, <*>, ~
Monad f        => (a -> f b) -> f a -> f b          Names: bind (flipped), flatMap, concatMap, =<<

flips

                   a -> (a -> b) -> b              Names: |>, #
Functor f     => f a -> (a -> b) -> f b            (never used???)
Applicative f => f a -> f (a -> b) -> f b          Names: <**> (rarely used???)
Monad f       => f a -> (a -> f b) -> f b          Names: bind, flatMap?, >>= (more common then its flip)
```