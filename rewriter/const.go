package rewriter

const (
	cstIterVar  = "ɪʇ" // it۰
	cstMoveNext = "MoveNext"
	cstCurrent  = "Current"

	cstYieldFromRangeVar = "ʌ" // v۰

	cstPairKey = "Key"
	cstPairVal = "Val"

	cstIterator       = "Iterator"
	cstNewStringIter  = "NewStringIter"
	cstNewIntegerIter = "NewIntegerIter"
	cstNewSliceIter   = "NewSliceIter"
	cstNewMapIter     = "NewMapIter"
	cstNewChanIter    = "NewChanIter"

	cstSeq      = "Seq"
	cstStart    = "Start"
	cstNormal   = "Normal"
	cstReturn   = "Return"
	cstBreak    = "Break"
	cstContinue = "Continue"
	cstDelay    = "Delay"
	cstBind     = "Bind"
	cstCombine  = "Combine"
	cstFor      = "For"
	cstRange    = "Range"
	cstLoop     = "Loop"
	cstWhile    = "While"
)

const (
	cstAPIReturnType = "Iter"
	cstAPIYield      = "Yield"
	cstAPIYieldFrom  = "YieldFrom"
)

const (
	// prevent name conflict for import name / pkg scope / local scope
	importSeqName = "ʂɘʠ" // seq۰
	importCoName  = "ɕɔ"  // co۰
	pkgCoName     = "co"
	pkgSeqName    = "seq"
	pkgCoPath     = "github.com/goghcrow/go-co"
	pkgSeqPath    = "github.com/goghcrow/go-co/seq"

	qualifiedIter      = pkgCoPath + "." + cstAPIReturnType
	qualifiedYield     = pkgCoPath + "." + cstAPIYield
	qualifiedYieldFrom = pkgCoPath + "." + cstAPIYieldFrom
)
