// golang library that creates a html file from a gdoc file
// author: prr
// created: 18/11/2021
// copyright 2022 prr, Peter Riemenschneider
//
// for changes see github
//
// start: CreGdocHtmlTil
//

package gdocHtml

import (
	"fmt"
	"os"
	"net/http"
	"io"
	"unicode/utf8"
	"google.golang.org/api/docs/v1"
)

const (
    PtTomm = 0.35277777777778
    MmTopt = 2.8346456692913207
)

type GdocHtmlObj struct {
	doc *docs.Document
	docName string
    docWidth float64
	docHeight float64
	ImgFoldName string
    ImgCount int
    tableCount int
    parCount int
	titleCount int
	subtitleCount int
    h1Count int
    h2Count int
    h3Count int
    h4Count int
    h5Count int
    h6Count int
	title namStyl
	subtitle namStyl
	h1 namStyl
	h2 namStyl
	h3 namStyl
	h4 namStyl
	h5 namStyl
	h6 namStyl
    spanCount int
	listStack *[]cList
	docLists []docList
	headings []heading
	sections []sect
	headCount int
	secCount int
	elCount int
	ftNoteCount int
	inImgCount int
	posImgCount int
	htmlFil *os.File
	folderName string
	folderPath string
    imgFoldNam string
    imgFoldPath string
	Options *OptObj
}

type dispObj struct {
	headCss string
	bodyHtml string
	bodyCss string
	tocHtml string
	tocCss string
}

type sect struct {
	sNum int
	secElStart int
	secElEnd int
}

type heading struct {
	id string
	hdElEnd int
	hdElStart int
}

type cList struct {
	cListId string
	cListCount int
	cOrd bool
}

type docList struct {
	listId string
	maxNestLev int64
	ord bool
}

type nestLevel struct {
	GlAl string
	GlFmt string
    GlSym string
	GlTyp string
	GlOrd bool
    Count int64
    FlInd float64
    StInd float64
	glTxtmap *textMap
}


type tabCell struct {
	pad [4] float64
	spad float64
	vert_align string
	bckcolor string
	border [4] tblBorder
	bwidth float64
	bdash string
	bcolor string
	cspan int
	rspan int
}

type tblBorder struct {
	color string
	dash string
	width float64
}

type namStyl struct {
	count int
	exist bool
}

type parMap struct {
	halign string
	headingId string
	direct bool
	indEnd float64
	indFlin float64
	indStart float64
	keepLines bool
	keepNext bool
	linSpac float64
	shading string
	spaceTop float64
	spaceBelow float64
	spaceMode string
	tabs []*tabStop
	pad [4]float64
	margin [4]float64
	hasBorders bool
	bordTop parBorder
	bordLeft parBorder
	bordRight parBorder
	bordBot parBorder
	bordBet parBorder
}

type parBorder struct {
	pad float64
	width float64
	color string
	dash string
}

type parStylRetObj struct {
	prefix string
	suffix string
	parId string
	cssStr string
}

type tabStop struct {
	tabAlign string
	offset float64
}

type textMap struct {
	bckColor string
	baseOffset string
	bold bool
	italic bool
	underline bool
	strike bool
	fontSize float64
	txtColor string
	link bool
	fontType string
	fontWeight	int64
}

type linkMap struct {
	url string
	bookmark string
}

type OptObj struct {
	DefLinSpacing float64
	BaseFontSize int
	CssFil bool
	ImgFold bool
    Verb bool
	Toc bool
	MultiSect bool
	DivBorders bool
	Divisions []string
	DocMargin [4]int
	ElMargin [4]int
}

func getDefOption(opt *OptObj) {

	opt.BaseFontSize = 0
	opt.MultiSect = false
	opt.DivBorders = false
	opt.DefLinSpacing = 1.2
	opt.DivBorders = false
	opt.CssFil = false
	opt.ImgFold = true
	opt.Verb = true
	opt.Toc = false
	for i:=0; i< 4; i++ {opt.ElMargin[i] = 0}

	opt.Divisions = []string{"Summary", "Main"}
	return
}

func printOptions (opt *OptObj) {

	fmt.Printf("\n************ Option Values ***********\n")
	fmt.Printf("  Base Font Size:       %d\n", opt.BaseFontSize)
	fmt.Printf("  Sections as <div>:    %t\n", opt.MultiSect)
	fmt.Printf("  Browser Line Spacing: %.1f\n",opt. DefLinSpacing)
	fmt.Printf("  <div> Borders:        %t\n", opt.DivBorders)
	fmt.Printf("  Divisions: %d\n", len(opt.Divisions))
	for i:=0; i < len(opt.Divisions); i++ {
		fmt.Printf("    div: %s\n", opt.Divisions[i])
	}
	fmt.Printf("  Separate CSS File:    %t\n", opt.CssFil)
	fmt.Printf("  Image Folder:         %t\n", opt.ImgFold)
	fmt.Printf("  Table of Content:     %t\n", opt.Toc)
	fmt.Printf("  Verbose output:       %t\n", opt.Verb)
	fmt.Printf("  Element Margin: ")
	for i:=0; i<4; i++ { fmt.Printf(" %3d",opt.ElMargin[i])}
	fmt.Printf("\n")
	fmt.Printf("***************************************\n\n")
}

func findDocList(list []docList, listid string) (res int) {

	res = -1
	for i:=0; i< len(list); i++ {
		if list[i].listId == listid {
			return i
		}
	}
	return res
}

func pushLiStack(stack *[]cList, item cList)(nstack *[]cList) {
	if (stack == nil) {
		xx := make([]cList,0)
		stack = &xx
	}
	*stack= append(*stack, item)
	return stack
}

func popLiStack(stack *[]cList)(nstack *[]cList) {

	if stack == nil {return nil}

	n := len(*stack) -1
	if n >= 0 {
		xx	:= (*stack)[:n]
		nstack = &xx
	} else {
		nstack = nil
	}
	return nstack
}

func getLiStack(stack *[]cList) (item cList, nl int) {

	if stack == nil {
		return item, -1
	}
	nl = len(*stack) -1
	// review n>1 if there is a  stack!
	if nl >= 0 {
		item = (*stack)[nl]
	} else {
		nl = -1
	}
	return item, nl
}

func printLiStack(stack *[]cList) {
var item cList
var	n int
	if stack == nil {
		fmt.Println("*** no listStack ***")
		return
	}
	n = len(*stack) -1
	if n>=0 {
		item = (*stack)[n]
	} else {
		n = -1
	}
	fmt.Printf("list stack Nlevel: %d", n)
	if n >= 0 {
		fmt.Printf(" id: %s ordered: %t", item.cListId, item.cOrd)
	}
	fmt.Printf("\n")
	return
}

func printLiStackItem(listAtt cList, cNest int){
		fmt.Printf("list Att Nlevel: %d", cNest)
		if cNest >= 0 {
			fmt.Printf(" id: %s ordered: %t", listAtt.cListId, listAtt.cOrd)
		}
		fmt.Printf("\n")
}

func getGlyphOrd(glyphTyp string)(bool) {
	ord := false
	switch glyphTyp {
		case "DECIMAL":
			ord = true
		case "ZERO_DECIMAL":
			ord = true
		case "UPPER_ALPHA":
			ord = true
		case "ALPHA":
			ord = true
		case "UPPER_ROMAN":
			ord = true
		case "ROMAN":
			ord = true
		default:
			ord = false
	}

	return ord
}

func printTxtMap(txtMap *textMap) {

	fmt.Println("********* text map ****************")
	fmt.Printf("Base Offset: %s\n", txtMap.baseOffset)
	fmt.Printf("Bold Text:   %t\n", txtMap.bold)
	fmt.Printf("Italic Text: %t\n", txtMap.italic)
	fmt.Printf("Underline:   %t\n", txtMap.underline)
	fmt.Printf("Text Strike: %t\n", txtMap.strike)
	fmt.Printf("Font:        %s\n", txtMap.fontType)
	fmt.Printf("Font Weight: %d\n", txtMap.fontWeight)
	fmt.Printf("Font Size:   %.1f\n", txtMap.fontSize)
	fmt.Printf("Font Color:  %s\n", txtMap.txtColor)
	fmt.Printf("Font BckCol: %s\n", txtMap.bckColor)

	return
}

func fillTxtMap(txtMap *textMap, txtStyl *docs.TextStyle)(alter bool, err error) {

	alter = false
	if txtStyl == nil {
		return alter, fmt.Errorf("decode txtstyle: -- no Style")
	}

	if txtStyl.BaselineOffset != txtMap.baseOffset {
		txtMap.baseOffset = txtStyl.BaselineOffset
		alter = true
	}

	if txtStyl.Bold != txtMap.bold {
		txtMap.bold = txtStyl.Bold
		alter = true
	}

	if txtStyl.Italic != txtMap.italic {
		txtMap.italic = txtStyl.Italic
		alter = true
	}

	if txtStyl.Underline != txtMap.underline {
 		txtMap.underline = txtStyl.Underline
		alter = true
	}

	if txtStyl.Strikethrough != txtMap.strike {
		txtMap.strike = txtStyl.Strikethrough
		alter = true
	}

	if txtStyl.WeightedFontFamily != nil {
		if txtStyl.WeightedFontFamily.FontFamily != txtMap.fontType {
			txtMap.fontType = txtStyl.WeightedFontFamily.FontFamily
			alter = true
		}
		if txtStyl.WeightedFontFamily.Weight != txtMap.fontWeight {
			txtMap.fontWeight = txtStyl.WeightedFontFamily.Weight
			alter = true
		}
	}
	if txtStyl.FontSize != nil {
		if txtStyl.FontSize.Magnitude != txtMap.fontSize {
			txtMap.fontSize = txtStyl.FontSize.Magnitude
			alter = true
		}
	}

	if txtStyl.ForegroundColor != nil {
		if txtStyl.ForegroundColor.Color != nil {
			color := getColor(txtStyl.ForegroundColor.Color)
			if color != txtMap.txtColor {
				txtMap.txtColor = color
				alter = true
			}
		}
	}

	if txtStyl.BackgroundColor != nil {
		if txtStyl.BackgroundColor.Color != nil {
			color := getColor(txtStyl.BackgroundColor.Color)
			if color != txtMap.bckColor {
				txtMap.bckColor = color
				alter = true
			}
		}
	}


	return alter, nil
}

func printParMap(parmap *parMap, parStyl *docs.ParagraphStyle) {

	alter := false

	if parStyl.Alignment != parmap.halign {
		fmt.Printf("align: %s %s \n", parmap.halign, parStyl.Alignment)
		parmap.halign = parStyl.Alignment
		alter = true
	}
	fmt.Printf("align: %s\n", parmap.halign)

	parmap.direct = true
	if (parStyl.IndentStart != nil) {
		if parStyl.IndentStart.Magnitude != parmap.indStart {
			fmt.Printf("indent start: %.1fpt %.1fpt\n", parmap.indStart, parStyl.IndentStart.Magnitude)
			parmap.indStart = parStyl.IndentStart.Magnitude
			alter = true
		}
	}
	fmt.Printf("indent start: %.1fpt\n", parmap.indStart)

	if (parStyl.IndentEnd != nil) {
		if parStyl.IndentEnd.Magnitude != parmap.indEnd {
			fmt.Printf("indent end: %.1f %.1f \n", parmap.indEnd, parStyl.IndentEnd.Magnitude)
			parmap.indEnd = parStyl.IndentEnd.Magnitude
			alter = true
		}
	}
	fmt.Printf("indent end: %.1fpt\n", parmap.indEnd)

	if (parStyl.IndentFirstLine != nil) {
		if parStyl.IndentFirstLine.Magnitude != parmap.indFlin {
			fmt.Printf("indent first line: %.1f %.1f \n", parmap.indFlin, parStyl.IndentFirstLine.Magnitude)
			parmap.indFlin = parStyl.IndentFirstLine.Magnitude
			alter = true
		}
	}
	fmt.Printf("indent first line: %.1fpt\n", parmap.indFlin)

	if parStyl.LineSpacing/100.0 != parmap.linSpac {
		fmt.Printf("line spacing: %.2f %.2f \n", parmap.linSpac, parStyl.LineSpacing/100.0)
		parmap.linSpac = parStyl.LineSpacing/100.0; alter = true;
	}
	fmt.Printf("line spacing: %.2fpt\n", parmap.linSpac)

	if parStyl.KeepLinesTogether != parmap.keepLines {
		fmt.Printf("keep Lines: %t %t\n", parmap.keepLines, parStyl.KeepLinesTogether)
		parmap.keepLines = parStyl.KeepLinesTogether; alter = true;
	}
	fmt.Printf("keep Lines: %t\n", parmap.keepLines)

	if parStyl.KeepWithNext != parmap.keepNext {
		fmt.Printf("keep With: %t %tf \n", parmap.keepNext, parStyl.KeepWithNext)
		parmap.keepNext = parStyl.KeepWithNext; alter = true;
	}
	fmt.Printf("keep With: %t\n", parmap.keepNext)

	if (parStyl.SpaceAbove != nil) {
		if parStyl.SpaceAbove.Magnitude != parmap.spaceTop {
			fmt.Printf("space above: %.1fpt %.1fpt\n", parmap.spaceTop, parStyl.SpaceAbove.Magnitude)
			parmap.spaceTop = parStyl.SpaceAbove.Magnitude
			alter = true
		}
	}
	fmt.Printf("space above: %.1fpt\n", parmap.spaceTop)

	if (parStyl.SpaceBelow != nil) {
		if parStyl.SpaceBelow.Magnitude != parmap.spaceBelow {
			fmt.Printf("space below: %.1f %.1f \n", parmap.spaceBelow, parStyl.SpaceBelow.Magnitude)
			parmap.spaceBelow = parStyl.SpaceBelow.Magnitude
			alter = true
		}
	}
	fmt.Printf("space below: %.1fpt\n", parmap.spaceBelow)

	if parStyl.SpacingMode != parmap.spaceMode {
		fmt.Printf("spacing mode: %s %s \n", parmap.spaceMode, parStyl.SpacingMode)
		parmap.spaceMode = parStyl.SpacingMode
		alter = true
	}
	fmt.Printf("spacing mode: %s\n", parmap.spaceMode)

	//tabs to do
//	parmap.hasBorders = true

	bb := true
	bb = bb && (parStyl.BorderBetween == nil)
	bb = bb && (parStyl.BorderTop == nil)
	bb = bb && (parStyl.BorderRight == nil)
	bb = bb && (parStyl.BorderBottom == nil)
	bb = bb && (parStyl.BorderLeft == nil)
	if bb {
		fmt.Printf("has no borders: %t %t \n", parmap.hasBorders, !bb)
		parmap.hasBorders = false
		fmt.Printf("alter 1: %t\n", alter)
		return
	}

	fmt.Println("has Borders!\n")
	parmap.hasBorders = true
	if parStyl.BorderBetween != nil {
		if parStyl.BorderBetween.Width != nil {
			if parStyl.BorderBetween.Width.Magnitude != parmap.bordBet.width {
				parmap.bordBet.width = parStyl.BorderBetween.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBetween.Padding != nil {
			if parStyl.BorderBetween.Padding.Magnitude != parmap.bordBet.pad {
				parmap.bordBet.pad = parStyl.BorderBetween.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBetween.Color != nil {
			if parStyl.BorderBetween.Color.Color != nil {
				color := getColor(parStyl.BorderBetween.Color.Color)
				if color != parmap.bordBet.color {
					parmap.bordBet.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderBetween.DashStyle != parmap.bordBet.dash {parmap.bordBet.dash = parStyl.BorderBetween.DashStyle; alter = true;}
	}

	if parStyl.BorderTop != nil {
		if parStyl.BorderTop.Width != nil {
			if parStyl.BorderTop.Width.Magnitude != parmap.bordTop.width {
				parmap.bordTop.width = parStyl.BorderTop.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderTop.Padding != nil {
			if parStyl.BorderTop.Padding.Magnitude != parmap.bordTop.pad {
				parmap.bordTop.pad = parStyl.BorderTop.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderTop.Color != nil {
			if parStyl.BorderTop.Color.Color != nil {
				color := getColor(parStyl.BorderTop.Color.Color)
				if color != parmap.bordTop.color {
					parmap.bordTop.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderTop.DashStyle != parmap.bordTop.dash {parmap.bordTop.dash = parStyl.BorderTop.DashStyle; alter = true;}
	}

	if parStyl.BorderRight != nil {
		if parStyl.BorderRight.Width != nil {
			if parStyl.BorderRight.Width.Magnitude != parmap.bordRight.width {
				parmap.bordRight.width = parStyl.BorderRight.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderRight.Padding != nil {
			if parStyl.BorderRight.Padding.Magnitude != parmap.bordRight.pad {
				parmap.bordRight.pad = parStyl.BorderRight.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderRight.Color != nil {
			if parStyl.BorderRight.Color.Color != nil {
				color := getColor(parStyl.BorderRight.Color.Color)
				if color != parmap.bordRight.color {
					parmap.bordRight.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderRight.DashStyle != parmap.bordRight.dash {
			parmap.bordRight.dash = parStyl.BorderRight.DashStyle
			alter = true
		}
	}

	if parStyl.BorderBottom != nil {
		if parStyl.BorderBottom.Width != nil {
			if parStyl.BorderBottom.Width.Magnitude != parmap.bordBot.width {
				parmap.bordBot.width = parStyl.BorderBottom.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBottom.Padding != nil {
			if parStyl.BorderBottom.Padding.Magnitude != parmap.bordBot.pad {
				parmap.bordBot.pad = parStyl.BorderBottom.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBottom.Color != nil {
			if parStyl.BorderBottom.Color.Color != nil {
				color := getColor(parStyl.BorderBottom.Color.Color)
				if color != parmap.bordBot.color {
					parmap.bordBot.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderBottom.DashStyle != parmap.bordBot.dash {parmap.bordBot.dash = parStyl.BorderBottom.DashStyle; alter = true;}
	}

	if parStyl.BorderLeft != nil {
		if parStyl.BorderLeft.Width != nil {
			if parStyl.BorderLeft.Width.Magnitude != parmap.bordLeft.width {
				parmap.bordLeft.width = parStyl.BorderLeft.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderLeft.Padding != nil {
			if parStyl.BorderLeft.Padding.Magnitude != parmap.bordLeft.pad {
				parmap.bordLeft.pad = parStyl.BorderLeft.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderLeft.Color != nil {
			if parStyl.BorderLeft.Color.Color != nil {
				color := getColor(parStyl.BorderLeft.Color.Color)
				if color != parmap.bordLeft.color {
					parmap.bordLeft.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderLeft.DashStyle != parmap.bordLeft.dash {parmap.bordLeft.dash = parStyl.BorderLeft.DashStyle; alter = true;}
	}

	bb2 := true
	bb2 = bb2 && (parmap.bordBet.width == 0.0)
	bb2 = bb2 && (parmap.bordTop.width == 0.0)
	bb2 = bb2 && (parmap.bordRight.width == 0.0)
	bb2 = bb2 && (parmap.bordBot.width == 0.0)
	bb2 = bb2 && (parmap.bordLeft.width == 0.0)

	if bb2 {parmap.hasBorders = false}

	fmt.Printf("alter borders: %t\n", alter)

	return
}


func fillParMap(parmap *parMap, parStyl *docs.ParagraphStyle)(alter bool, err error) {

	alter = false
	if parStyl == nil {
		return alter, fmt.Errorf("no parStyl!")
	}

	if parStyl.Alignment != parmap.halign {
//fmt.Printf("align: %s : %s \n", parmap.halign,parStyl.Alignment)
		if len(parStyl.Alignment) > 0 {
			if !(len(parmap.halign)>0) {alter =true}
			parmap.halign = parStyl.Alignment
		}
	}
	parmap.direct = true

	if (parStyl.IndentStart != nil) {
		if parStyl.IndentStart.Magnitude != parmap.indStart {
			parmap.indStart = parStyl.IndentStart.Magnitude
			alter = true
		}
	}
	if (parStyl.IndentEnd != nil) {
		if parStyl.IndentEnd.Magnitude != parmap.indEnd {
			parmap.indEnd = parStyl.IndentEnd.Magnitude
			alter = true
		}
	}
	if (parStyl.IndentFirstLine != nil) {
		if parStyl.IndentFirstLine.Magnitude != parmap.indFlin {
			parmap.indFlin = parStyl.IndentFirstLine.Magnitude
			alter = true
		}
	}

	if parStyl.LineSpacing/100.0 != parmap.linSpac {
// fmt.Printf("line spacing: %.2f %.2f\n", parmap.linSpac, parStyl.LineSpacing/100.0)
		if parStyl.LineSpacing > 1.0 {
			parmap.linSpac = parStyl.LineSpacing/100.0
			alter = true
		}
	}

	// may have to introduce an exemption for title
	if !parmap.keepLines {
		if parStyl.KeepLinesTogether != parmap.keepLines {
			parmap.keepLines = parStyl.KeepLinesTogether
			alter = true
		}
	}

	if !parmap.keepNext {
		if parStyl.KeepWithNext != parmap.keepNext {
			parmap.keepNext = parStyl.KeepWithNext
			alter =true
		}
	}

	if (parStyl.SpaceAbove != nil) {
		if parStyl.SpaceAbove.Magnitude != parmap.spaceTop {
			parmap.spaceTop = parStyl.SpaceAbove.Magnitude
			alter = true
		}
	}
	if (parStyl.SpaceBelow != nil) {
		if parStyl.SpaceBelow.Magnitude != parmap.spaceBelow {
			parmap.spaceBelow = parStyl.SpaceBelow.Magnitude
			alter = true
		}
	}


	if parStyl.SpacingMode != parmap.spaceMode {
		if (len(parStyl.SpacingMode) > 0) {
			parmap.spaceMode = parStyl.SpacingMode
			alter = true
		}
	}

//fmt.Printf("fillParMap 1: %t\n", alter)
//fmt.Printf("fillParMap 2: %t\n", alter)

	//tabs to do
//	parmap.hasBorders = true

	// check for zero width borders
	bb := true
	if (parStyl.BorderBetween != nil) {
		if parStyl.BorderBetween.Width != nil {
			if parStyl.BorderBetween.Width.Magnitude > 0.1 {
				bb = false
			}
		}
	}

	if (parStyl.BorderTop != nil) {
		if parStyl.BorderTop.Width != nil {
			if parStyl.BorderTop.Width.Magnitude > 0.1 {
				bb = false
			}
		}
	}

	if parStyl.BorderRight != nil {
		if parStyl.BorderRight.Width != nil {
			if parStyl.BorderRight.Width.Magnitude > 0.1 {
				bb = false
			}
		}
	}

	if parStyl.BorderBottom != nil {
		if parStyl.BorderBottom.Width != nil {
			if parStyl.BorderBottom.Width.Magnitude > 0.1 {
				bb = false
			}
		}
	}


	if parStyl.BorderLeft != nil {
		if parStyl.BorderLeft.Width != nil {
			if parStyl.BorderLeft.Width.Magnitude > 0.1 {
				bb = false
			}
		}
	}

	if bb {
		parmap.hasBorders = false
//fmt.Printf("no border return: %t\n", alter)
//fmt.Printf("fillParMap 3: %t\n", alter)
		return alter, nil
	}


	parmap.hasBorders = true
	if parStyl.BorderBetween != nil {
		if parStyl.BorderBetween.Width != nil {
			if parStyl.BorderBetween.Width.Magnitude != parmap.bordBet.width {
				parmap.bordBet.width = parStyl.BorderBetween.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBetween.Padding != nil {
			if parStyl.BorderBetween.Padding.Magnitude != parmap.bordBet.pad {
				parmap.bordBet.pad = parStyl.BorderBetween.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBetween.Color != nil {
			if parStyl.BorderBetween.Color.Color != nil {
				color := getColor(parStyl.BorderBetween.Color.Color)
				if color != parmap.bordBet.color {
					parmap.bordBet.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderBetween.DashStyle != parmap.bordBet.dash {parmap.bordBet.dash = parStyl.BorderBetween.DashStyle; alter = true;}
	}

	if parStyl.BorderTop != nil {
		if parStyl.BorderTop.Width != nil {
			if parStyl.BorderTop.Width.Magnitude != parmap.bordTop.width {
				parmap.bordTop.width = parStyl.BorderTop.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderTop.Padding != nil {
			if parStyl.BorderTop.Padding.Magnitude != parmap.bordTop.pad {
				parmap.bordTop.pad = parStyl.BorderTop.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderTop.Color != nil {
			if parStyl.BorderTop.Color.Color != nil {
				color := getColor(parStyl.BorderTop.Color.Color)
				if color != parmap.bordTop.color {
					parmap.bordTop.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderTop.DashStyle != parmap.bordTop.dash {parmap.bordTop.dash = parStyl.BorderTop.DashStyle; alter = true;}
	}

	if parStyl.BorderRight != nil {
		if parStyl.BorderRight.Width != nil {
			if parStyl.BorderRight.Width.Magnitude != parmap.bordRight.width {
				parmap.bordRight.width = parStyl.BorderRight.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderRight.Padding != nil {
			if parStyl.BorderRight.Padding.Magnitude != parmap.bordRight.pad {
				parmap.bordRight.pad = parStyl.BorderRight.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderRight.Color != nil {
			if parStyl.BorderRight.Color.Color != nil {
				color := getColor(parStyl.BorderRight.Color.Color)
				if color != parmap.bordRight.color {
					parmap.bordRight.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderRight.DashStyle != parmap.bordRight.dash {
			parmap.bordRight.dash = parStyl.BorderRight.DashStyle
			alter = true
		}
	}

	if parStyl.BorderBottom != nil {
		if parStyl.BorderBottom.Width != nil {
			if parStyl.BorderBottom.Width.Magnitude != parmap.bordBot.width {
				parmap.bordBot.width = parStyl.BorderBottom.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBottom.Padding != nil {
			if parStyl.BorderBottom.Padding.Magnitude != parmap.bordBot.pad {
				parmap.bordBot.pad = parStyl.BorderBottom.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderBottom.Color != nil {
			if parStyl.BorderBottom.Color.Color != nil {
				color := getColor(parStyl.BorderBottom.Color.Color)
				if color != parmap.bordBot.color {
					parmap.bordBot.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderBottom.DashStyle != parmap.bordBot.dash {parmap.bordBot.dash = parStyl.BorderBottom.DashStyle; alter = true;}
	}

	if parStyl.BorderLeft != nil {
		if parStyl.BorderLeft.Width != nil {
			if parStyl.BorderLeft.Width.Magnitude != parmap.bordLeft.width {
				parmap.bordLeft.width = parStyl.BorderLeft.Width.Magnitude
				alter = true
			}
		}
		if parStyl.BorderLeft.Padding != nil {
			if parStyl.BorderLeft.Padding.Magnitude != parmap.bordLeft.pad {
				parmap.bordLeft.pad = parStyl.BorderLeft.Padding.Magnitude
				alter = true
			}
		}
		if parStyl.BorderLeft.Color != nil {
			if parStyl.BorderLeft.Color.Color != nil {
				color := getColor(parStyl.BorderLeft.Color.Color)
				if color != parmap.bordLeft.color {
					parmap.bordLeft.color = color
					alter = true
				}
			}
		}
		if parStyl.BorderLeft.DashStyle != parmap.bordLeft.dash {parmap.bordLeft.dash = parStyl.BorderLeft.DashStyle; alter = true;}
	}

	bb2 := true
	bb2 = bb2 && (parmap.bordBet.width == 0.0)
	bb2 = bb2 && (parmap.bordTop.width == 0.0)
	bb2 = bb2 && (parmap.bordRight.width == 0.0)
	bb2 = bb2 && (parmap.bordBot.width == 0.0)
	bb2 = bb2 && (parmap.bordLeft.width == 0.0)

	if bb2 {parmap.hasBorders = false}


	return alter, nil
}

func cvtTxtMapCss(txtMap *textMap)(cssStr string) {

	cssStr =""
	if len(txtMap.baseOffset) > 0 {
		switch txtMap.baseOffset {
			case "SUPERSCRIPT":
				cssStr += "  vertical-align: sub;\n"
			case "SUBSCRIPT":
				cssStr += "	vertical-align: sup;\n"
			case "NONE":

			default:
			//error
				cssStr += fmt.Sprintf("/* Baseline Offset unknown: %s */\n", txtMap.baseOffset)
		}
	}
	if txtMap.bold {
		cssStr += "  font-weight: bold;\n"
	}
	if txtMap.italic { cssStr += "  font-style: italic;\n"}
	if txtMap.underline { cssStr += "  text-decoration: underline;\n"}
	if txtMap.strike { cssStr += "  text-decoration: line-through;\n"}

	if len(txtMap.fontType) >0 { cssStr += fmt.Sprintf("  font-family: %s;\n", txtMap.fontType)}
	if txtMap.fontWeight > 0 {cssStr += fmt.Sprintf("  font-weight: %d;\n", txtMap.fontWeight)}
	if txtMap.fontSize >0 {cssStr += fmt.Sprintf("  font-size: %.2fpt;\n", txtMap.fontSize)}
	if len(txtMap.txtColor) >0 {cssStr += fmt.Sprintf("  color: %s;\n", txtMap.txtColor)}
	if len(txtMap.bckColor) >0 {cssStr += fmt.Sprintf("  background-color: %s;\n", txtMap.bckColor)}

	return cssStr
}

func addDispObj(src, add *dispObj) {
	src.headCss += add.headCss
	src.bodyHtml += add.bodyHtml
	src.bodyCss += add.bodyCss
	src.tocHtml += add.tocHtml
	src.tocCss += add.tocCss
	return
}

func getColor(color  *docs.Color)(outstr string) {
    outstr = ""
        if color != nil {
            blue := int(color.RgbColor.Blue*255.0)
            red := int(color.RgbColor.Red*255.0)
            green := int(color.RgbColor.Green*255)
            outstr += fmt.Sprintf("rgb(%d, %d, %d)", red, green, blue)
            return outstr
        }
    outstr = "/*no color*/\n"
    return outstr
}

func getDash(dashStyle string)(outstr string) {

	switch dashStyle {
		case "SOLID":
			outstr = "solid"
		case "DOT":
			outstr = "dotted"
		case "DASH":
			outstr = "dashed"
		default:
			outstr = "none"
	}

    return outstr
}

func getImgLayout (layout string) (ltyp int, err error) {

	switch layout {
		case "WRAP_TEXT":

		case "BREAK_LEFT":

		case "BREAK_RIGHT":

		case "BREAK_LEFT_RIGHT":

		case "IN_FRONT_OF_TEXT":

		case "BEHIND_TEXT":

		default:
			return -1, fmt.Errorf("layout %s not implemented!", layout)
	}
	return ltyp, nil
}

func tcell_vert_align (alStr string) (outstr string) {
	switch alStr {
		case "TOP":
			outstr = "top"
		case "Middle":
		 	outstr = "middle"
		case "BOTTOM":
			outstr = "bottom"
		default:
			outstr = "baseline"
	}
	return outstr
}

func creHtmlHead()(outstr string, err error) {
	outstr = "<!DOCTYPE html>\n"
//	outstr += fmt.Sprintf("<!-- file: %s -->\n", dObj.docName)
	outstr += "<head>\n<style>\n"
	return outstr, nil
}


func (dObj *GdocHtmlObj) printHeadings() {

	if len(dObj.headings) == 0 {
		fmt.Println("*** no Headings ***")
		return
	}

	fmt.Printf("**** Headings: %d ****\n", len(dObj.headings))
	for i:=0; i< len(dObj.headings); i++ {
		fmt.Printf("  heading %3d  Id: %-15s El Start:%3d End:%3d\n", i, dObj.headings[i].id,
			dObj.headings[i].hdElStart, dObj.headings[i].hdElEnd)
	}
}

func (dObj *GdocHtmlObj) cvtParMapCss(pMap *parMap)(cssStr string) {
	cssStr =""

	if len(pMap.halign) > 0 {
		switch pMap.halign {
			case "START":
				cssStr += "  text-align: left;\n"
			case "CENTER":
				cssStr += "  text-align: center;\n"
			case "END":
				cssStr += "  text-align: right;\n"
			case "JUSTIFIED":
				cssStr += "  text-align: justify;\n"
			default:
				cssStr += fmt.Sprintf("/* unrecognized Alignment %s */\n", pMap.halign)
		}

	}

	opt := dObj.Options
	if pMap.linSpac > 0.0 {
		if opt.DefLinSpacing > 0.0 {
			cssStr += fmt.Sprintf("  line-height: %.2f;\n", opt.DefLinSpacing*pMap.linSpac)
		} else {
			cssStr += fmt.Sprintf("  line-height: %.2f;\n", pMap.linSpac)
		}
	}

	if pMap.indFlin > 0.0 {
		cssStr += fmt.Sprintf("  text-indent: %.1fpt;\n", pMap.indFlin)
	}

	mlCss :=""
	if pMap.indStart > 0.0 {
		mlCss = fmt.Sprintf("%.1fpt", pMap.indStart)
	} else {
		mlCss = fmt.Sprintf("0")
	}
	mrCss:=""
	if pMap.indEnd > 0.0 {
		mrCss = fmt.Sprintf("%.1fpt", pMap.indEnd)
	} else {
		mrCss = fmt.Sprintf("0")
	}
	mtCss := ""
	if pMap.spaceTop > 0.0 {
		mtCss = fmt.Sprintf("%.1fpt", pMap.spaceTop)
	} else {
		mtCss = fmt.Sprintf("0")
	}
	mbCss := ""
	if pMap.spaceBelow > 0.0 {
		mbCss = fmt.Sprintf("%.1fpt", pMap.spaceBelow)
	} else {
		mbCss = fmt.Sprintf("0")
	}

	cssStr += fmt.Sprintf("  margin: %s %s %s %s;\n", mtCss, mrCss, mbCss, mlCss)

	if !pMap.hasBorders { return cssStr }
	cssStr += fmt.Sprintf("  padding: %.1fpt %.1fpt %.1fpt %.1fpt;\n", pMap.bordTop.pad, pMap.bordRight.pad, pMap.bordBot.pad, pMap.bordLeft.pad)
	cssStr += fmt.Sprintf("  border-top: %.1fpt %s %s;\n", pMap.bordTop.width, getDash(pMap.bordTop.dash), pMap.bordTop.color)
	cssStr += fmt.Sprintf("  border-right: %.1fpt %s %s;\n", pMap.bordRight.width, getDash(pMap.bordRight.dash), pMap.bordRight.color)
	cssStr += fmt.Sprintf("  border-bottom: %.1fpt %s %s;\n", pMap.bordBot.width, getDash(pMap.bordBot.dash), pMap.bordBot.color)
	cssStr += fmt.Sprintf("  border-left: %.1fpt %s %s;\n", pMap.bordLeft.width, getDash(pMap.bordLeft.dash), pMap.bordLeft.color)

	return cssStr
}

func (dObj *GdocHtmlObj) getNamedStyl(namedTyp string)(parStyl *docs.ParagraphStyle, txtStyl *docs.TextStyle, err error) {
	var namStyl *docs.NamedStyle

	doc:= dObj.doc
// initialise named styles
	namStyles := doc.NamedStyles
	stylIdx := -1

// find normal style first
	for istyl:=0; istyl<len(namStyles.Styles); istyl++ {
		if namStyles.Styles[istyl].NamedStyleType == namedTyp {
			stylIdx = istyl
			namStyl = namStyles.Styles[istyl]
			break
		}
	}

	if stylIdx < 0 {
		return nil, nil, fmt.Errorf("cannot find named style %s!", namedTyp)
	}

	parStyl = namStyl.ParagraphStyle
	txtStyl = namStyl.TextStyle
	return parStyl, txtStyl, nil
}


func (dObj *GdocHtmlObj) downloadImg()(err error) {

    doc := dObj.doc
	verb := dObj.Options.Verb
    if !(len(dObj.imgFoldNam) >0) {
        return fmt.Errorf("no ingFoldNam!")
    }
    imgFoldPath := dObj.imgFoldPath + "/"
    fmt.Println("image folder: ", imgFoldPath)

    fmt.Printf("*** Inline Imgs: %d ***\n", len(doc.InlineObjects))
    for k, inlObj := range doc.InlineObjects {
        imgProp := inlObj.InlineObjectProperties.EmbeddedObject.ImageProperties
        if verb {
            fmt.Printf("Source: %s Obj %s\n", k, imgProp.SourceUri)
            fmt.Printf("Content: %s Obj %s\n", k, imgProp.ContentUri)
        }
        if !(len(imgProp.SourceUri) > 0) {
            return fmt.Errorf("image %s has no URI\n", k)
        }
        imgNam := imgFoldPath + k[4:] + ".jpeg"
        if verb {fmt.Printf("image path: %s\n", imgNam)}
        URL := imgProp.ContentUri
        httpResp, err := http.Get(URL)
        if err != nil {
            return fmt.Errorf("hhtp.Get: could not fetch %s! %v", URL, err)
        }
        defer httpResp.Body.Close()
//  fmt.Printf("http got %s!\n", URL)
        if httpResp.StatusCode != 200 {
            return fmt.Errorf("httpResp: Received non 200 response code %d!", httpResp.StatusCode)
        }
//  fmt.Printf("http status: %d\n ", httpResp.StatusCode)
    //Create a empty file
        outfil, err := os.Create(imgNam)
        if err != nil {
            return fmt.Errorf("os.Create - cannot create img file! %v", err)
        }
        defer outfil.Close()
//  fmt.Println("created dir")
        //Write the bytes to the fiel
        _, err = io.Copy(outfil, httpResp.Body)
        if err != nil {
            return fmt.Errorf("io.Copy: %v", err)
        }
    }


    fmt.Printf("*** Positioned Imgs: %d ***\n", len(doc.PositionedObjects))
    for k, posObj := range doc.PositionedObjects {
        imgProp := posObj.PositionedObjectProperties.EmbeddedObject.ImageProperties
        if verb {
            fmt.Printf("Source: %s Obj %s\n", k, imgProp.SourceUri)
            fmt.Printf("Content: %s Obj %s\n", k, imgProp.ContentUri)
        }
        if !(len(imgProp.SourceUri) > 0) {
            return fmt.Errorf("image %s has no URI\n", k)
        }
        imgNam := imgFoldPath + k[4:] + ".jpeg"
        if verb {fmt.Printf("image path: %s\n", imgNam)}
        URL := imgProp.ContentUri
        httpResp, err := http.Get(URL)
        if err != nil {
            return fmt.Errorf("http.Get: could not fetch img: %s! %v", URL, err)
        }
        defer httpResp.Body.Close()
//  fmt.Printf("http got %s!\n", URL)
        if httpResp.StatusCode != 200 {
            return fmt.Errorf("httpResp: Received non 200 response code %d!", httpResp.StatusCode)
        }
//  fmt.Printf("http status: %d\n ", httpResp.StatusCode)
    //Create a empty file


        outfil, err := os.Create(imgNam)
        if err != nil {
            return fmt.Errorf("os.Create: cannot create img file! %v", err)
        }
        defer outfil.Close()
//  fmt.Println("created dir")
        //Write the bytes to the fiel
        _, err = io.Copy(outfil, httpResp.Body)
        if err != nil {
            return fmt.Errorf("io.Copy: cannot copy img file content! %v", err)
        }
    }

    return nil
}

func (dObj *GdocHtmlObj) createOutFil(divNam string) (err error) {

	if len(divNam) > 0 {
		fnam := dObj.docName + "_" + divNam
	} else {
		fnam := dObj.docName
	}
	filpath := dObj.folderPath + "/" + fnam + ".html"
	fmt.Printf("filpath: %s fnam: %s\n", fnam, filpath)
	outfil, err := os.Create(filpath)
	if err != nil {
		return fmt.Errorf("os.Create: cannot create html file: %v", err)
	}
	dObj.htmlFil = outfil
	return nil
}


func (dObj *GdocHtmlObj) createHtmlFolder(path string)(err error) {
	var filnam, foldnam string

	if len(path) == 0 {
        return fmt.Errorf("folder path name %s non-existant!", path)
	}
	// check whether html folder exists
    filnam = dObj.docName
    if len(filnam) < 2 {
        return fmt.Errorf("docName %s too short!", filnam)
    }

	lenPath := len(path)
	if lenPath > 0 {
		if path[lenPath -1] != '/' { path += "/"}
		foldnam = path + filnam
	} else {
		foldnam = filnam
	}

    // check whether dir folder exists, if not create one
    newDir := false
    _, err = os.Stat(foldnam)
    if os.IsNotExist(err) {
        err1 := os.Mkdir(foldnam, os.ModePerm)
        if err1 != nil {
            return fmt.Errorf("os.Mkdir: could not create html folder! %v", err1)
        }
        newDir = true
    } else {
        if err != nil {
            return fmt.Errorf("os.Stat: could not find html folder! %v", err)
        }
    }

    // open directory
    if !newDir {
        err = os.RemoveAll(foldnam)
        if err != nil {
            return fmt.Errorf("os.RemoveAll: could not delete files in html folder! %v", err)
        }
        err = os.Mkdir(foldnam, os.ModePerm)
        if err != nil {
            return fmt.Errorf("os.Mkdir:: could not create html folder! %v", err)
        }
    }

    dObj.folderName = filnam
    dObj.folderPath = foldnam

    return nil
}

func (dObj *GdocHtmlObj) createImgFolder()(err error) {

    filnam :=dObj.docName
    if len(filnam) < 2 {
        return fmt.Errorf("filename %s too short!", filnam)
    }

    bf := []byte(filnam)
    // replace empty space with underscore
    for i:= 0; i< len(filnam); i++ {
        if bf[i] == ' ' {bf[i]='_'}
        if bf[i] == '.' {
            return fmt.Errorf("parameter 'filnam' has period!")
        }
    }

    imgFoldNam := "imgs_" + string(bf)

//    fmt.Println("output file name: ", dObj.htmlFil.Name())
    foldNamb := []byte(dObj.htmlFil.Name())
    idx := 0
    for i:=len(foldNamb)-1; i> 0; i-- {
        if foldNamb[i] == '/' {
            idx = i
            break
        }
    }

    imgFoldPath := imgFoldNam
    if idx > 0 {
        imgFoldPath = string(foldNamb[:idx]) + "/" + imgFoldNam
    }

    fmt.Println("img folder path: ", imgFoldPath)

    // check whether dir folder exists, if not create one
    newDir := false
    _, err = os.Stat(imgFoldPath)
    if os.IsNotExist(err) {
        err1 := os.Mkdir(imgFoldPath, os.ModePerm)
        if err1 != nil {
            return fmt.Errorf("os.Mkdir: could not create img folder! %v", err1)
        }
        newDir = true
    } else {
        if err != nil {
            return fmt.Errorf("os.Stat: could not find img folder! %v", err)
        }
    }

    // open directory
    if !newDir {
        err = os.RemoveAll(imgFoldPath)
        if err != nil {
            return fmt.Errorf("os.RemoveAll: could not delete files in image folder! %v", err)
        }
        err = os.Mkdir(imgFoldPath, os.ModePerm)
        if err != nil {
            return fmt.Errorf("os.Mkdir: could not create img folder! %v", err)
        }
    }
    dObj.imgFoldNam = imgFoldNam
    dObj.imgFoldPath = imgFoldPath

    return nil
}

func (dObj *GdocHtmlObj) disp_GdocHtmlObj (dbgfil *os.File) (err error) {
	var outstr string

  	if dObj == nil {
        return fmt.Errorf("dObj is nil!")
    }
	if dbgfil == nil {
        return fmt.Errorf("dbgfil is nil!")
    }

	outstr = fmt.Sprintf("Document: %s\n", dObj.docName)
	outstr += "Lists:\n"

	dbgfil.WriteString(outstr)
	return nil
}


func (dObj *GdocHtmlObj) findListProp (listId string) (listProp *docs.ListProperties) {

	found := false
	doc := dObj.doc

	for key, listItem := range doc.Lists  {
		if listId == key {
			listProp = listItem.ListProperties
			found = true
			break
		}
	}

	if found { return listProp}

	return nil
}
/*
func (dObj *GdocHtmlObj) initGdocFolder(doc *docs.Document, path string) (err error) {
	if doc == nil {
		return fmt.Errorf("no doc provided!")
	}

	// need to transform file name
	// replace spaces with underscore
	dNam := doc.Title
	x := []byte(dNam)
	for i:=0; i<len(x); i++ {
		if x[i] == ' ' {
			x[i] = '_'
		}
	}
	dObj.docName = string(x[:])

//ggg
	err = dObj.createHtmlFolder(path)
	if err!= nil {
		return fmt.Errorf("createHtmlFolder: %v", err)
	}

	return nil
}
*/

func (dObj *GdocHtmlObj) initGdocHtml(doc *docs.Document, options *OptObj) (err error) {
	var listItem docList
	var heading heading
	var sec sect

//	if options == nil {return fmt.Errorf("no option!")}
	if doc == nil {return fmt.Errorf("no doc provided!")}
	dObj.doc = doc

	// need to transform file name
	// replace spaces with underscore
	dNam := doc.Title
	x := []byte(dNam)
	for i:=0; i<len(x); i++ {
		if x[i] == ' ' {
			x[i] = '_'
		}
	}
	dObj.docName = string(x[:])

	if options == nil {
		defOpt := new(OptObj)
		getDefOption(defOpt)
		if defOpt.Verb {printOptions(defOpt)}
		dObj.Options = defOpt
	} else {
		dObj.Options = options
	}


// initialise heading use
// each heading has a default paragraph style
// no reason to  create css if heading is not used
	dObj.title.exist = false
	dObj.subtitle.exist = false
	dObj.h1.exist = false
	dObj.h2.exist = false
	dObj.h3.exist = false
	dObj.h4.exist = false
	dObj.h5.exist = false
	dObj.h6.exist = false

// section breaks
	dObj.elCount = len(doc.Body.Content)
	dObj.ftNoteCount = 0
	parHdEnd := 0
	// last element of section
	secPtEnd := 0
	// set up first page
	sec.secElStart = 0
	dObj.sections = append(dObj.sections, sec)
	seclen := len(dObj.sections)
//		fmt.Println("el: ", el, "section len: ", seclen, " secPtEnd: ", secPtEnd)

	for el:=0; el<dObj.elCount; el++ {
		elObj:= doc.Body.Content[el]
		if elObj.SectionBreak != nil {
			if elObj.SectionBreak.SectionStyle.SectionType == "NEXT_PAGE" {
//sss
				sec.secElStart = el
				dObj.sections = append(dObj.sections, sec)
				seclen := len(dObj.sections)
//		fmt.Println("el: ", el, "section len: ", seclen, " secPtEnd: ", secPtEnd)
				if seclen > 1 {
					dObj.sections[seclen-2].secElEnd = secPtEnd
				}
			}
		}
		// paragraphs and lists
		if elObj.Paragraph != nil {

			if elObj.Paragraph.Bullet != nil {
				// lists
				listId := elObj.Paragraph.Bullet.ListId
				found := findDocList(dObj.docLists, listId)
				if found < 0 {
					listItem.listId = listId
					listItem.maxNestLev = elObj.Paragraph.Bullet.NestingLevel
					dObj.docLists = append(dObj.docLists, listItem)
				} else {
					nestlev := elObj.Paragraph.Bullet.NestingLevel
					if dObj.docLists[found].maxNestLev < nestlev { dObj.docLists[found].maxNestLev = nestlev }
				}
			}
			// headings
			if len(elObj.Paragraph.ParagraphStyle.HeadingId) > 0 {
				heading.id = elObj.Paragraph.ParagraphStyle.HeadingId
				heading.hdElStart = el
				dObj.headings = append(dObj.headings, heading)
				hdlen := len(dObj.headings)
//		fmt.Println("el: ", el, "hdlen: ", hdlen, "parHdEnd: ", parHdEnd)
				if hdlen > 1 {
					dObj.headings[hdlen-2].hdElEnd = parHdEnd
				}
			}

			// footnotes
			for parEl:=0; parEl<len(elObj.Paragraph.Elements); parEl++ {
				parElObj := elObj.Paragraph.Elements[parEl]
				if parElObj.FootnoteReference != nil {dObj.ftNoteCount++}
			}
			parHdEnd = el
			secPtEnd = el
		} // end paragraph
	} // end el loop

	hdlen := len(dObj.headings)
	if hdlen > 0 {
		dObj.headings[hdlen-1].hdElEnd = parHdEnd
	}
	seclen = len(dObj.sections)
	if seclen > 0 {
		dObj.sections[seclen-1].secElEnd = secPtEnd
	}

	if dObj.Options.Verb {
		fmt.Printf("********** Headings in Document: %2d ***********\n", len(dObj.headings))
		for i:=0; i< len(dObj.headings); i++ {
			fmt.Printf("  heading %3d  Id: %-15s El Start:%3d End:%3d\n", i, dObj.headings[i].id, dObj.headings[i].hdElStart, dObj.headings[i].hdElEnd)
		}
		fmt.Printf("\n********** Pages in Document: %2d ***********\n", len(dObj.sections))
		for i:=0; i< len(dObj.sections); i++ {
			fmt.Printf("  Page %3d  El Start:%3d End:%3d\n", i, dObj.sections[i].secElStart, dObj.sections[i].secElEnd)
		}
		fmt.Printf("\n************ Lists in Document: %2d ***********\n", len(dObj.docLists))
		for i:=0; i< len(dObj.docLists); i++ {
			fmt.Printf("list %3d id: %s max level: %d ordered: %t\n", i, dObj.docLists[i].listId, dObj.docLists[i].maxNestLev, dObj.docLists[i].ord)
		}
		fmt.Printf("***********************************************\n\n")
	}

// Headings

// images
	dObj.inImgCount = len(doc.InlineObjects)
	dObj.posImgCount = len(doc.PositionedObjects)

//    dObj.parCount = len(doc.Body.Content)

	return nil
}

func (dObj *GdocHtmlObj) dlImages()(err error) {
// function that creates image folder and downloads images
    totObjNum := dObj.inImgCount + dObj.posImgCount
	if totObjNum > 0{
		if dObj.Options.Verb {fmt.Printf("*** creating image folder for %d images ***\n", totObjNum)}

    	err = dObj.createImgFolder()
		if err != nil {
        	return fmt.Errorf("dObj.createImgFolder: could not create ImgFolder: %v!", err)
		}
		if dObj.Options.Verb {
			fmt.Printf("Created Image folder: %s\n", dObj.ImgFoldName)
		}
    	err = dObj.downloadImg()
    	if err != nil {
        	return fmt.Errorf("downloadImg: could not download images: %v!", err)
    	}
	} else {
		if dObj.Options.Verb {fmt.Printf("***** no image folder created *****\n")}
	}
	if dObj.Options.Verb {fmt.Printf("**************************************\n\n")}
	return nil
}


func (dObj *GdocHtmlObj) cvtGlyph(nlev *docs.NestingLevel)(cssStr string) {

	var glyphTyp string

	// ordered list
		switch nlev.GlyphType {
			case "DECIMAL":
				glyphTyp = "decimal"
			case "ZERO_DECIMAL":
				glyphTyp = "decimal-leading-zero"

			case "ALPHA":
				glyphTyp = "lower-alpha"

 			case "UPPER_ALPHA":
				glyphTyp = "upper-alpha"

			case "ROMAN":
				glyphTyp = "lower-roman"

			case "UPPER_ROMAN":
				glyphTyp = "upper-roman"

			default:
//				cssStr = "/* unknown GlyphType */\n"
			glyphTyp = ""
		}
		if len(glyphTyp) > 0 {
			cssStr = "  list-style-type: " + glyphTyp +";\n"
		} else {
	// unordered list
		cssStr =fmt.Sprintf("/*-Glyph Symbol:%x - */\n",nlev.GlyphSymbol)
		r, _ := utf8.DecodeRuneInString(nlev.GlyphSymbol)

		switch r {
			case 9679:
				glyphTyp = "disc"

			case 9675:
				glyphTyp = "circle"

			case 9632:
				glyphTyp = "square"

			default:

		}
		if len(glyphTyp) > 0 {
			cssStr = "  list-style-type: " + glyphTyp +";\n"
//			cssStr +="  list-style-position: inside;\n"
//			cssStr +="  padding-left: 0;\n"
		}
	}

	return cssStr
}

func (dObj *GdocHtmlObj) cvtInlineImg(imgEl *docs.InlineObjectElement)(htmlStr string, cssStr string, err error) {

	if imgEl == nil {
		return "","", fmt.Errorf("cvtInlineImg:: imgEl is nil!")
	}
	doc := dObj.doc

	imgElId := imgEl.InlineObjectId
	if !(len(imgElId) > 0) {return "","", fmt.Errorf("cvtInlineImg:: no InlineObjectId found!")}

	// need to remove first part of the id
	idx := 0
	for i:=0; i< len(imgElId); i++ {
		if imgElId[i] == '.' {
			idx = i+1
			break
		}
	}
	imgId :=""
	if (idx>0) && (idx<len(imgElId)-1) {
		imgId = "img_" + imgElId[idx:]
	}

	// need to change for imagefolder
	htmlStr = fmt.Sprintf("<!-- inline image %s -->\n", imgElId)
	imgObj := doc.InlineObjects[imgElId].InlineObjectProperties.EmbeddedObject

	if dObj.Options.ImgFold {
    	imgSrc := dObj.imgFoldNam + "/" + imgId + ".jpeg"
		htmlStr +=fmt.Sprintf("<img src=\"%s\" id=\"%s\" alt=\"%s\">\n", imgSrc, imgId, imgObj.Title)
	} else {
		htmlStr +=fmt.Sprintf("<img src=\"%s\" id=\"%s\" alt=\"%s\">\n", imgObj.ImageProperties.SourceUri, imgId, imgObj.Title)
	}
	cssStr = fmt.Sprintf("#%s {\n",imgId)
	cssStr += fmt.Sprintf(" width:%.1fpt; height:%.1fpt; \n}\n", imgObj.Size.Width.Magnitude, imgObj.Size.Height.Magnitude )
	// todo add margin
	return htmlStr, cssStr, nil
}

func (dObj *GdocHtmlObj) cvtParElText(parElTxt *docs.TextRun)(htmlStr string, cssStr string, err error) {

   if parElTxt == nil {
        return "","", fmt.Errorf("cvtPelText -- parElTxt is nil!")
    }

	// need to check whether <1
	if len(parElTxt.Content) < 2 { return "","",nil}
	// need to compare text style with the default style
	spanCssStr, err := dObj.cvtTxtStylCss(parElTxt.TextStyle, false)
	if err != nil {
		spanCssStr = fmt.Sprintf("/*error parEl Css %v*/\n", err) + spanCssStr
	}
	linkPrefix := ""
	linkSuffix := ""
	if parElTxt.TextStyle.Link != nil {
		if len(parElTxt.TextStyle.Link.Url)>0 {
			linkPrefix = "<a href = \"" + parElTxt.TextStyle.Link.Url + "\">"
			linkSuffix = "</a>"
		}
	}
	if len(spanCssStr)>0 {
		dObj.spanCount++
		spanIdStr := fmt.Sprintf("%s_sp%d", dObj.docName, dObj.spanCount)
		cssStr = fmt.Sprintf("#%s {\n", spanIdStr) + spanCssStr + "}\n"
		htmlStr = fmt.Sprintf("<span id=\"%s\">",spanIdStr) + linkPrefix + parElTxt.Content + linkSuffix + "</span>"
	} else {
		htmlStr = linkPrefix + parElTxt.Content + linkSuffix
	}
	return htmlStr, cssStr, nil
}


func (dObj *GdocHtmlObj) closeList(nl int)(htmlStr string) {
	// ends a list

	if (dObj.listStack == nil) {return ""}

	stack := dObj.listStack
	n := len(*stack) -1

	for i := n; i > nl; i-- {
		ord := (*stack)[i].cOrd
		if ord {
			htmlStr += "</ol>\n"
		} else {
			htmlStr +="</ul>\n"
		}
		nstack := popLiStack(stack)
		dObj.listStack = nstack
	}
	return htmlStr
}

func (dObj *GdocHtmlObj) renderPosImg(posImg docs.PositionedObject, posId string)(htmlStr, cssStr string, err error) {

	posObjProp := posImg.PositionedObjectProperties
	imgProp := posObjProp.EmbeddedObject
	htmlStr += fmt.Sprintf("\n<!-- Positioned Image %s -->\n", posId)
	imgDivId := fmt.Sprintf("%s_%s", dObj.docName, posId[4:])
	imgId := imgDivId + "_img"
	pimgId := imgDivId +"_p"

	layout := posObjProp.Positioning.Layout
	topPos := posObjProp.Positioning.TopOffset.Magnitude
	leftPos := posObjProp.Positioning.LeftOffset.Magnitude
	fmt.Printf("layout %s top: %.1fmm left:%.1fmm\n", layout, topPos*PtTomm, leftPos*PtTomm)

	imgSrc := imgProp.ImageProperties.ContentUri
	if dObj.Options.ImgFold {
		imgSrc = dObj.imgFoldNam + "/" + posId[4:] + ".jpeg"
	}

	switch layout {
		case "WRAP_TEXT", "BREAK_LEFT":
			cssStr += fmt.Sprintf("#%s {\n", imgId)
			cssStr += fmt.Sprintf("float:left; clear:both;")
			cssStr += fmt.Sprintf("  width:%.1fpt; height:%.1fpt;\n",imgProp.Size.Width.Magnitude, imgProp.Size.Height.Magnitude)
			cssStr += fmt.Sprintf("  margin: %.1fpt %.1fpt %.1fpt %.1fpt;\n", imgProp.MarginTop.Magnitude, imgProp.MarginRight.Magnitude, imgProp.MarginBottom.Magnitude, imgProp.MarginLeft.Magnitude)
			cssStr += "}\n"
			cssStr += fmt.Sprintf("#%s {\n", pimgId)
			cssStr += fmt.Sprintf("  margin-left: %.1fpt; margin-right: %.1fpt;\n", imgProp.MarginLeft.Magnitude, imgProp.MarginRight.Magnitude)
			cssStr += "}\n"
			cssStr += fmt.Sprintf("#%s:before {\n", imgDivId)
			cssStr += fmt.Sprintf("content:''; display:block; float:left; height:%.1fmm;\n",topPos*PtTomm)
			cssStr += "}\n"

		case "BREAK_RIGHT":
			cssStr += fmt.Sprintf("#%s {\n", imgId)
			cssStr += fmt.Sprintf("float:right; clear:both;")
			cssStr += fmt.Sprintf("  width:%.1fpt; height:%.1fpt;\n",imgProp.Size.Width.Magnitude, imgProp.Size.Height.Magnitude)
			cssStr += fmt.Sprintf("  margin: %.1fpt %.1fpt %.1fpt %.1fpt;\n", imgProp.MarginTop.Magnitude, imgProp.MarginRight.Magnitude, imgProp.MarginBottom.Magnitude, imgProp.MarginLeft.Magnitude)
			cssStr += "}\n"
			cssStr += fmt.Sprintf("#%s {\n", pimgId)
			cssStr += fmt.Sprintf("  margin-left: %.1fpt; margin-right: %.1fpt;\n", imgProp.MarginLeft.Magnitude, imgProp.MarginRight.Magnitude)
			cssStr += "}\n"
			cssStr += fmt.Sprintf("#%s:before {\n", imgDivId)
			cssStr += fmt.Sprintf("content:''; display:block; float:right; height:%.1fmm;\n",topPos*PtTomm)
			cssStr += "}\n"

		case "BREAK_LEFT_RIGHT":

		case "IN_FRONT_OF_TEXT":
// absolute
		case "BEHIND_TEXT":
// absolute
		default:
			cssStr += fmt.Sprintf("#%s {\n", imgId)
			cssStr += fmt.Sprintf("  width:%.1fpt; height:%.1fpt;\n",imgProp.Size.Width.Magnitude, imgProp.Size.Height.Magnitude)
			cssStr += fmt.Sprintf("  margin: %.1fpt %.1fpt %.1fpt %.1fpt;\n", imgProp.MarginTop.Magnitude, imgProp.MarginRight.Magnitude, imgProp.MarginBottom.Magnitude, imgProp.MarginLeft.Magnitude)
			cssStr += "}\n"
			cssStr += fmt.Sprintf("#%s {\n", pimgId)
			cssStr += fmt.Sprintf("  margin-left: %.1fpt; margin-right: %.1fpt;\n", imgProp.MarginLeft.Magnitude, imgProp.MarginRight.Magnitude)
			cssStr += "}\n"
	}

	htmlStr += fmt.Sprintf("  <div id=\"%s\">\n",imgDivId)
	htmlStr += fmt.Sprintf("     <img src=\"%s\" alt=\"%s\" id=\"%s\">\n", imgSrc, imgProp.Title, imgId)
//	htmlStr += fmt.Sprintf("     <p id=\"%s\">%s</p>\n", pimgId, imgProp.Title)
	htmlStr += "  </div>\n"

	return htmlStr, cssStr, nil
}

// table element
// 	tObj, _ := dObj.cvtTableToHtml(tableEl)

func (dObj *GdocHtmlObj) cvtTable(tbl *docs.Table)(tabObj dispObj, err error) {
	var htmlStr, cssStr string
	var tabWidth float64
	var icol, trow int64
	var defcel tabCell


	doc := dObj.doc
	dObj.tableCount++
//	tblId := fmt.Sprintf("%s_tab_%d", dObj.docName, dObj.tableCount)

    docPg := doc.DocumentStyle
    PgWidth := docPg.PageSize.Width.Magnitude
    NetPgWidth := PgWidth - (docPg.MarginLeft.Magnitude + docPg.MarginRight.Magnitude)
//   fmt.Printf("Default Table Width: %.1f", NetPgWidth)
    tabWidth = NetPgWidth
	tabw := 0.0
    for icol=0; icol < tbl.Columns; icol++ {
        tcolObj :=tbl.TableStyle.TableColumnProperties[icol]
        if tcolObj.Width != nil {
            tabw += tbl.TableStyle.TableColumnProperties[icol].Width.Magnitude
        }
    }
	if tabw > 0.0 {tabWidth = tabw}

// table cell default values
// define default cell classs
	tcelDef := tbl.TableRows[0].TableCells[0]
	tcelDefStyl := tcelDef.TableCellStyle

// default values which google does not set but uses
	defcel.vert_align = "top"
	defcel.bcolor = "black"
	defcel.bwidth = 1.0
	defcel.bdash = "solid"

// xxx
	if tcelDefStyl != nil {
		defcel.vert_align = tcell_vert_align(tcelDefStyl.ContentAlignment)

// if left border is the only border specified, let's use it for default values
		tb := (tcelDefStyl.BorderTop == nil)&& (tcelDefStyl.BorderRight == nil)
		tb = tb&&(tcelDefStyl.BorderBottom == nil)
		if (tcelDefStyl.BorderLeft != nil) && tb {
			if tcelDefStyl.BorderLeft.Color != nil {defcel.border[3].color = getColor(tcelDefStyl.BorderLeft.Color.Color)}
			if tcelDefStyl.BorderLeft.Width != nil {defcel.border[3].width = tcelDefStyl.BorderLeft.Width.Magnitude}
			defcel.border[3].dash = getDash(tcelDefStyl.BorderLeft.DashStyle)
		}

		if tcelDefStyl.PaddingTop != nil {defcel.pad[0] = tcelDefStyl.PaddingTop.Magnitude}
		if tcelDefStyl.PaddingRight != nil {defcel.pad[1] = tcelDefStyl.PaddingRight.Magnitude}
		if tcelDefStyl.PaddingBottom != nil {defcel.pad[2] = tcelDefStyl.PaddingBottom.Magnitude}
		if tcelDefStyl.PaddingLeft != nil {defcel.pad[3] = tcelDefStyl.PaddingLeft.Magnitude}

		if tcelDefStyl.BackgroundColor != nil {defcel.bckcolor = getColor(tcelDefStyl.BackgroundColor.Color)}

		if tcelDefStyl.BorderTop != nil {
			if tcelDefStyl.BorderTop.Color != nil {defcel.border[0].color = getColor(tcelDefStyl.BorderTop.Color.Color)}
			if tcelDefStyl.BorderTop.Width != nil {defcel.border[0].width = tcelDefStyl.BorderTop.Width.Magnitude}
			defcel.border[0].dash = getDash(tcelDefStyl.BorderTop.DashStyle)
		}
		if tcelDefStyl.BorderRight != nil {
			if tcelDefStyl.BorderRight.Color != nil {defcel.border[1].color = getColor(tcelDefStyl.BorderRight.Color.Color)}
			if tcelDefStyl.BorderRight.Width != nil {defcel.border[1].width = tcelDefStyl.BorderRight.Width.Magnitude}
			defcel.border[1].dash = getDash(tcelDefStyl.BorderRight.DashStyle)
		}
		if tcelDefStyl.BorderBottom != nil {
			if tcelDefStyl.BorderBottom.Color != nil {defcel.border[2].color = getColor(tcelDefStyl.BorderBottom.Color.Color)}
			if tcelDefStyl.BorderBottom.Width != nil {defcel.border[2].width = tcelDefStyl.BorderBottom.Width.Magnitude}
			defcel.border[2].dash = getDash(tcelDefStyl.BorderBottom.DashStyle)
		}
		if tcelDefStyl.BorderLeft != nil {
			if tcelDefStyl.BorderLeft.Color != nil {defcel.border[3].color = getColor(tcelDefStyl.BorderLeft.Color.Color)}
			if tcelDefStyl.BorderLeft.Width != nil {defcel.border[3].width = tcelDefStyl.BorderLeft.Width.Magnitude}
			defcel.border[3].dash = getDash(tcelDefStyl.BorderLeft.DashStyle)
		}
		if tcelDefStyl.BorderTop == tcelDefStyl.BorderRight {
//			fmt.Println("same border!")
			if tcelDefStyl.BorderTop != nil {
				if tcelDefStyl.BorderTop.Color != nil {defcel.bcolor = getColor(tcelDefStyl.BorderTop.Color.Color)}
				defcel.bdash = getDash(tcelDefStyl.BorderTop.DashStyle)
				if tcelDefStyl.BorderTop.Width != nil {defcel.bwidth = tcelDefStyl.BorderTop.Width.Magnitude}
			}
		}
	}

	//set up table
	tblClass := fmt.Sprintf("%s_tbl", dObj.docName)
	tblCellClass := fmt.Sprintf("%s_tcel", dObj.docName)
	htmlStr = ""

	// if there is an open list, close it
	if len(*dObj.listStack) >= 0 {
		htmlStr += dObj.closeList(len(*dObj.listStack))
fmt.Printf("table closing list!\n")
	}

	htmlStr += fmt.Sprintf("<table class=\"%s\">\n", tblClass)

  // table styling
  	cssStr = fmt.Sprintf(".%s {\n",tblClass)
 	cssStr += fmt.Sprintf("  border: 1px solid black;\n  border-collapse: collapse;\n")
 	cssStr += fmt.Sprintf("  width: %.1fpt;\n", tabWidth)
	cssStr += "   margin:auto;\n"
	cssStr += "}\n"

// table columns
	tabWtyp :=tbl.TableStyle.TableColumnProperties[0].WidthType
//fmt.Printf("table width type: %s\n", tabWtyp)
	if tabWtyp == "FIXED_WIDTH" {
		htmlStr +="<colgroup>\n"
		for icol = 0; icol < tbl.Columns; icol++ {
			colId := fmt.Sprintf("tab%d_col%d", dObj.tableCount, icol)
			cssStr += fmt.Sprintf("#%s {width: %.1fpt;}\n", colId, tbl.TableStyle.TableColumnProperties[icol].Width.Magnitude)
			htmlStr += fmt.Sprintf("<col span=\"1\" id=\"%s\">\n", colId)
		}
		htmlStr +="</colgroup>\n"
	}


	cssStr += fmt.Sprintf(".%s {\n",tblCellClass)
 	cssStr += fmt.Sprintf("  border: %.1fpt %s %s;\n", defcel.bwidth, defcel.bdash, defcel.bcolor)
	cssStr += fmt.Sprintf("  vertical-align: %s;\n", defcel.vert_align )
	cssStr += fmt.Sprintf("  padding: %.1fpt %.1fpt %.1fpt %.1fpt;\n", defcel.pad[0], defcel.pad[1], defcel.pad[2], defcel.pad[3])
 	cssStr += "}\n"


// row styling
	htmlStr += "  <tbody>\n"
	tblCellCount := 0
	for trow=0; trow < tbl.Rows; trow++ {
		htmlStr += fmt.Sprintf("  <tr>\n")
		trowobj := tbl.TableRows[trow]
//		mrheight := trowobj.TableRowStyle.MinRowHeight.Magnitude

		numCols := len(trowobj.TableCells)
		for tcol:=0; tcol< numCols; tcol++ {
			tcell := trowobj.TableCells[tcol]
			tblCellCount++
			cellStr := ""
			celId := fmt.Sprintf("tab%d_cell%d", dObj.tableCount, tblCellCount)
			// check whether cell style is different from default
			if tcell.TableCellStyle != nil {
				tstyl := tcell.TableCellStyle
				if tstyl.BackgroundColor != nil {cellStr += fmt.Sprintf(" background-color:\"%s\";",getColor(tstyl.BackgroundColor.Color))}
				if tcell_vert_align(tstyl.ContentAlignment) != defcel.vert_align {cellStr += fmt.Sprintf(" vertical-align: %s;", tcell_vert_align(tstyl.ContentAlignment))}
				if tstyl.PaddingTop != nil {
					if tstyl.PaddingTop.Magnitude != defcel.pad[0] { cellStr += fmt.Sprintf(" padding-top: %5.1fpt;", tstyl.PaddingTop.Magnitude)}
				}

				if tstyl.BorderTop != nil {
					// Color
					if tstyl.BorderTop.Color != nil {cellStr += fmt.Sprintf(" border-top-color: %s;", getColor(tstyl.BorderTop.Color.Color))}
					//dash
					if getDash(tstyl.BorderTop.DashStyle) != defcel.bdash {cellStr += fmt.Sprintf(" border-top-style: %s;",  getDash(tstyl.BorderTop.DashStyle))}
					//Width
					if tstyl.BorderTop.Width != nil {cellStr += fmt.Sprintf(" border-top-width: %5.1fpt;", tstyl.BorderTop.Width.Magnitude)}
				}
			}
// xxxx
			if len(cellStr) >0 {
				cssStr += fmt.Sprintf("#%s {",celId)
				cssStr += fmt.Sprintf("%s }\n", cellStr)
				htmlStr += fmt.Sprintf("    <td id=\"%s\" class=\"%s\">\n", celId, tblCellClass)
			} else {
			// default
				htmlStr += fmt.Sprintf("    <td class=\"%s\">\n", tblCellClass)
			}
			elNum := len(tcell.Content)
			for el:=0; el< elNum; el++ {
				elObj := tcell.Content[el]
				tObj, err:=dObj.cvtContentEl(elObj)
				if err != nil {
					tabObj.bodyHtml = htmlStr
					tabObj.bodyCss = cssStr
					return tabObj, fmt.Errorf("ConvertTable: %v", err)
				}
				cssStr += tObj.bodyCss
				htmlStr += "    " + tObj.bodyHtml
			}
			htmlStr += "  </td>\n"

		}
		htmlStr += "</tr>\n"
	}

	htmlStr += "  </tbody>\n</table>\n"
	tabObj.bodyHtml = htmlStr
	tabObj.bodyCss = cssStr
	return tabObj, nil
}

func (dObj *GdocHtmlObj) cvtPar(par *docs.Paragraph)(parObj dispObj, err error) {
// paragraph element par
// - Bullet
// - Elements
// - ParagraphStyle
// - Positioned Objects
//
	var parHtmlStr, parCssStr string
	var prefix, suffix string
	var tocPrefix, tocSuffix string
	var listPrefix, listHtml, listCss, listSuffix string
	var newList cList

	if par == nil {
        return parObj, fmt.Errorf("cvtPar -- parEl is nil!")
    }
	if dObj == nil {
        return parObj, fmt.Errorf("cvttPar -- dObj is nil!")
    }
	errStr := ""
	dObj.parCount++

	parHtmlStr = ""
	parCssStr = ""

	isList := false
	if par.Bullet != nil {isList = true}
//fmt.Printf("********** par %d list: %t ***********\n", dObj.parCount, isList)

	if par.Bullet == nil {
		// if there was an open list, close it
		if dObj.listStack != nil {
			parHtmlStr += dObj.closeList(-1)
//fmt.Printf("par bullet close list\n")
		}
	}

// Positioned Objects
	numPosObj := len(par.PositionedObjectIds)
	for i:=0; i< numPosObj; i++ {
		posId := par.PositionedObjectIds[i]
		posObj, ok := dObj.doc.PositionedObjects[posId]
		if !ok {return parObj, fmt.Errorf("cvtPar: could not find positioned Object with id: ", posId)}

		imgHtmlStr, imgCssStr, err := dObj.renderPosImg(posObj, posId)
		if err != nil {
			parHtmlStr += fmt.Sprintf("<!-- error cvtPar:: render pos img %v -->\n", err) + imgHtmlStr
			parCssStr += imgCssStr
		} else {
			parHtmlStr += imgHtmlStr
			parCssStr += imgCssStr
		}
	}

	parObj.bodyHtml += parHtmlStr
	parObj.bodyCss += parCssStr
	// need to reset
	parHtmlStr = ""
	parCssStr = ""

	// check for new line paragraph
		if len(par.Elements) == 1 {
			if par.Elements[0].TextRun != nil {
				if par.Elements[0].TextRun.Content == "\n" {
					parObj.bodyHtml = "<br>\n"
					return parObj, nil
				}
			}
		}


	namedTyp := par.ParagraphStyle.NamedStyleType
	namParStyl, _, err := dObj.getNamedStyl(namedTyp)
	if err != nil {
		return parObj, fmt.Errorf("cvtPar: %v", err)
	}

	// default style for each named style
	// add css for named style at the begining of the Css
	// normal_text is already defined as the default in the css for the <div>
	// *** important *** cvtNamedStyl needs to be run before CvtParStyle 
	if namedTyp != "NORMAL_TEXT" {
		hdcss, err := dObj.cvtNamedStyl(namedTyp)
		if err != nil {
			errStr = fmt.Sprintf("%v", err)
		}
		parObj.headCss += hdcss + errStr
	}

	if par.Bullet == nil {
		// now we have a normal paragraph element
		parHtmlStr += fmt.Sprintf("\n<!-- Paragraph %d %s -->\n", dObj.parCount, namedTyp)
	}


//	namParStyl, _, _ = dObj.getNamedStyl(namedTyp)

//zz
	parStylCss :=""
	parStylCss, prefix, suffix, err = dObj.cvtParStyl(par.ParagraphStyle, namParStyl, isList)
	if err != nil {
		errStr = fmt.Sprintf("/* error cvtParStyl: %v */\n",err)
	}
//fmt.Printf("par %d:  %s %s %s\n", dObj.parCount, prefix, suffix, namedTyp)
	parObj.bodyCss += errStr + parStylCss

	// Heading Id refers to a heading not just a normal paragraph
	hdHtmlStr:=""
	if len(par.ParagraphStyle.HeadingId) > 0 {
		hdHtmlStr = fmt.Sprintf("<!-- Heading Id: %s -->", par.ParagraphStyle.HeadingId)
	}
	if len(hdHtmlStr) > 0 {parHtmlStr += hdHtmlStr + "\n"}


	decode := true
	errStr = ""

	// par elements: text and css for text
	numParEl := len(par.Elements)
    for pEl:=0; pEl< numParEl; pEl++ {
        parEl := par.Elements[pEl]
		elHtmlStr, elCssStr, err := dObj.cvtParEl(parEl)
		if err != nil { parHtmlStr += fmt.Sprintf("<!-- error cvtParEl: %v -->\n",err)}
      	parHtmlStr += elHtmlStr
		parCssStr += elCssStr

	} // loop par el

// lists
    if par.Bullet != nil {

		// there is paragraph style for each ul and a text style for each list element
		txtmap := new(textMap)
		if par.Bullet.TextStyle != nil {
			_, err := fillTxtMap(txtmap,par.Bullet.TextStyle)
			if err != nil { return parObj, fmt.Errorf("cvtPar List getting text style %v", err)}

		}

		if dObj.Options.Verb {listHtml += fmt.Sprintf("<!-- List Element %d -->\n", dObj.parCount)}

		// find list id of paragraph
		listid := par.Bullet.ListId
		nestIdx := int(par.Bullet.NestingLevel)

		// retrieve the list properties from the doc.Lists map
		listProp := dObj.doc.Lists[listid].ListProperties
		glyphTyp := listProp.NestingLevels[nestIdx].GlyphType
		listOrd := getGlyphOrd(glyphTyp)

		// A. check whether need new <ul> or <ol>
		// listHtml contains the <ul> <ol> element
		listHtml = ""
//		listSid := listid[4:]

		// conditions for new <ul><ol>
		// 1. beginning of a list
		// 2. increase in nesting level
		// 3. different listid -> old list ended; beginning of new list

		// condition for </ul></ol>
		// 1. decrease in nesting level

//		fmt.Println("*********** listStack **********")
//		fmt.Printf("listid: %s \n", listid)
//		printLiStack(dObj.listStack)

		listAtt, cNest := getLiStack(dObj.listStack)
//		printLiStackItem(listAtt, cNest)
//lll
		switch listid == listAtt.cListId {
			case true:
				switch {
					case nestIdx > cNest:
						for nl:=cNest; nl < nestIdx; nl++ {
							newList.cListId = listid
							newList.cOrd = listOrd
							newStack := pushLiStack(dObj.listStack, newList)
							dObj.listStack = newStack
							if listOrd {
								listHtml = fmt.Sprintf("<ol class=\"%s_ul nL_%d\">\n", listid[4:], nl)
							} else {
								listHtml = fmt.Sprintf("<ul class=\"%s_ul nL_%d\">\n", listid[4:], nl)
							}
						}
				listHtml += fmt.Sprintf("<!-- same list increase %s new NL %d  old Nl %d -->\n", listid, nestIdx, cNest)
//				fmt.Printf("<!-- same list increase %s new NL %d  old Nl %d -->\n", listid, nestIdx, cNest)
//	printLiStack(dObj.listStack)

					case nestIdx < cNest:
/*
						for nl:=cNest; nl>= nestIdx; nl++ {
							newStack := popLiStack(dObj.listStack)
							dObj.listStack = newStack
						}
*/
						listHtml = dObj.closeList(nestIdx)
				listHtml += fmt.Sprintf("<!-- same list reduce %s new NL %d  old Nl %d -->\n", listid, nestIdx, cNest)
//				fmt.Printf("<!-- same list reduce %s new NL %d  old Nl %d -->\n", listid, nestIdx, cNest)
					case nestIdx == cNest:
						listHtml =""
				}

			case false:
				// new list
				// close list first
				listHtml = dObj.closeList(-1)
				listHtml += fmt.Sprintf("<!-- new list %s %s -->\n", listid, listAtt.cListId)
//			fmt.Printf("<!-- new list %s %s -->\n", listid, listAtt.cListId)
				newList.cListId = listid
				newList.cOrd = listOrd
				newStack := pushLiStack(dObj.listStack, newList)
				dObj.listStack = newStack
				if listOrd {
					listHtml += fmt.Sprintf("<ol class=\"%s_ul nL_%d\">\n", listid[4:], nestIdx)
				} else {
					listHtml += fmt.Sprintf("<ul class=\"%s_ul nL_%d\">\n", listid[4:], nestIdx)
				}
			// if no cList list -> Clistid == ""


		}
		// need to add mark
		//listPrefix := fmt.Sprintf("<li class=\"%s_li nL%d mk_%d\">\n"
		listPrefix = fmt.Sprintf("<li class=\"%s_li nL_%d\">", listid[4:], nestIdx)

		listSuffix = "</li>"

	}

	parObj.bodyCss += listCss + parCssStr
	parObj.bodyHtml += listHtml + listPrefix + prefix + parHtmlStr + suffix + listSuffix + "\n"
	if decode {
		parObj.tocHtml += tocPrefix + parHtmlStr + tocSuffix + "\n"
	}
	return parObj, nil
}

func (dObj *GdocHtmlObj) cvtParEl(parEl *docs.ParagraphElement)(htmlStr string, cssStr string, err error) {

		if parEl.InlineObjectElement != nil {
        	imgHtmlStr, imgCssStr, err := dObj.cvtInlineImg(parEl.InlineObjectElement)
        	if err != nil {
            	htmlStr += fmt.Sprintf("<!-- error cvtPelToHtml: %v -->\n",err)
        	}
        	htmlStr += imgHtmlStr
			cssStr += imgCssStr
		}

		if parEl.TextRun != nil {
        	txtHtmlStr, txtCssStr, err := dObj.cvtParElText(parEl.TextRun)
        	if err != nil {
            	htmlStr += fmt.Sprintf("<!-- error cvtPelToHtml: %v -->\n",err)
        	}
        	htmlStr += txtHtmlStr
			cssStr += txtCssStr
		}

        if parEl.HorizontalRule != nil {

        }

        if parEl.ColumnBreak != nil {

        }

        if parEl.Person != nil {

        }

        if parEl.RichLink != nil {

        }

	return htmlStr, cssStr, nil
}

func (dObj *GdocHtmlObj) cvtNamedStyl(namedStylTyp string)(cssStr string, err error) {

	cssComment:=""
	namParStyl, namTxtStyl, err := dObj.getNamedStyl(namedStylTyp)
	if err != nil {
		cssComment = fmt.Sprintf("  /* cvtNamedStyle: named Style not recognized */\n")
		return cssComment, nil
	}


	parmap := new(parMap)
	txtmap := new(textMap)

	_, err = fillParMap(parmap, namParStyl)
	if err != nil {
		cssComment = fmt.Sprintf("  /* cvtNamedStyle: error fillParMap */\n")
		return cssComment, nil
	}

	_, err = fillTxtMap(txtmap, namTxtStyl)
	if err != nil {
		cssComment = fmt.Sprintf("  /* cvtNamedStyle: named Style not recognized */\n")
		return cssComment, nil
	}

	cssPrefix := ""
	switch namedStylTyp {
		case "TITLE":
			if !(dObj.title.exist)  {
				cssPrefix = fmt.Sprintf(".%s_title {\n", dObj.docName)
				dObj.title.exist = true
			}

		case "SUBTITLE":
			if !(dObj.subtitle.exist) {
				cssPrefix =fmt.Sprintf(".%s_subtitle {\n",dObj.docName)
				dObj.subtitle.exist = true
			}
		case "HEADING_1":
			if !(dObj.h1.exist) {
				cssPrefix =fmt.Sprintf(".%s_h1 {\n",dObj.docName)
				dObj.h1.exist = true
			}
		case "HEADING_2":
			if !(dObj.h2.exist) {
				cssPrefix =fmt.Sprintf(".%s_h2 {\n",dObj.docName)
				dObj.h2.exist = true
			}
		case "HEADING_3":
			if !(dObj.h3.exist) {
				cssPrefix =fmt.Sprintf(".%s_h3 {\n",dObj.docName)
				dObj.h3.exist = true
			}
		case "HEADING_4":
			if !(dObj.h4.exist) {
				cssPrefix =fmt.Sprintf(".%s_h4 {\n",dObj.docName)
				dObj.h4.exist = true
			}
		case "HEADING_5":
			if !(dObj.h5.exist) {
				cssPrefix =fmt.Sprintf(".%s_h5 {\n",dObj.docName)
				dObj.h5.exist = true
			}
		case "HEADING_6":
			if !(dObj.h6.exist) {
				cssPrefix =fmt.Sprintf(".%s_h6 {\n",dObj.docName)
				dObj.h6.exist = true
			}
		case "NORMAL_TEXT":

		case "NAMED_STYLE_TYPE_UNSPECIFIED":

		default:

	}
	if len(cssPrefix) > 0 {
		parCss := dObj.cvtParMapCss(parmap)
		txtCss := cvtTxtMapCss(txtmap)
		cssStr += cssPrefix + parCss + txtCss + "}\n"
	}
	return cssStr, nil
}

func (dObj *GdocHtmlObj) cvtParStyl(parStyl, namParStyl *docs.ParagraphStyle, isList bool)(cssStr, prefix, suffix string, err error) {

	cssComment:=""
	if namParStyl == nil {
		// def error the default is that the normal_text paragraph style is passed
		cssComment = fmt.Sprintf("/* Paragraph Style: no named Style */\n")
		return cssComment, "", "", nil
	}

	cssComment = fmt.Sprintf("/* Paragraph Style: %s */\n", parStyl.NamedStyleType )

	alter:= false
	parmap := new(parMap)
	cssParAtt := ""

	_, err = fillParMap(parmap, namParStyl)
	if err != nil {
		cssComment += "/* erro fill Parmap namparstyl */" + fmt.Sprintf("%v\n", err)
	}

// fmt.Printf("begin fillparmap parstyl %s: %t\n", parStyl.NamedStyleType, alter)

	if parStyl == nil || isList {
		cssParAtt = dObj.cvtParMapCss(parmap)
	} else {
		alter, err = fillParMap(parmap, parStyl)
		if err != nil {
			cssComment += "/* erro fill Parmap parstyl */" + fmt.Sprintf("%v\n", err)
		}
		if alter {cssParAtt = dObj.cvtParMapCss(parmap)}
	}
 //fmt.Printf("*** parstyle %s alter: %t\n", parStyl.NamedStyleType, alter)

//ppp
//	printParMap(parmap, parStyl)
	// NamedStyle Type
	isListClass := ""
	if isList {
		isListClass = " list"
	}

	prefix = ""
	suffix = ""
	cssPrefix := ""
	headingId := parStyl.HeadingId

	switch parStyl.NamedStyleType {
		case "TITLE":
			if dObj.title.exist && !alter {
				prefix = fmt.Sprintf("<p class=\"%s_title%s\"", dObj.docName, isListClass)
			}
			if alter {
				cssPrefix = fmt.Sprintf(".%s_title_%d {\n",dObj.docName, dObj.titleCount)
				prefix = fmt.Sprintf("<p class=\"%s_title_%d\"", dObj.docName, dObj.titleCount)
				dObj.titleCount++
			}
			suffix = "</p>"

		case "SUBTITLE":
			if dObj.subtitle.exist && !alter {
				prefix = fmt.Sprintf("<p class=\"%s_subtitle\"", dObj.docName)
			}
			if alter {
				cssPrefix = fmt.Sprintf(".s_subtitle_%d {\n",dObj.docName, dObj.subtitleCount)
				prefix = fmt.Sprintf("<p class=\"%s_subtitle_%d\"", dObj.docName, dObj.subtitleCount)
				dObj.subtitleCount++
			}
			suffix = "</p>"

		case "HEADING_1":
			if dObj.h1.exist && !alter {
				prefix = fmt.Sprintf("<h1 class=\"%s_h1\"", dObj.docName)
			}
			if alter {
				cssPrefix = fmt.Sprintf(".%s_h1_%d {\n",dObj.docName, dObj.h1Count)
				prefix = fmt.Sprintf("<h1 class=\"%s_h1_%d\"", dObj.docName, dObj.h1Count)
				dObj.h1Count++
			}
			suffix = "</h1>"
		case "HEADING_2":
			if dObj.h2.exist && !alter {
				prefix = fmt.Sprintf("<h2 class=\"%s_h2\"", dObj.docName)
			}
			if alter {
				cssPrefix = fmt.Sprintf(".%s_h2_%d {\n",dObj.docName, dObj.h2Count)
				prefix = fmt.Sprintf("<h2 class=\"%s_h2_%d\"", dObj.docName, dObj.h2Count)
				dObj.h2Count++
			}
			suffix = "</h2>"
		case "HEADING_3":
			if dObj.h3.exist && !alter {
				prefix = fmt.Sprintf("<h3 class=\"%s_h3\"", dObj.docName)
			}
			if alter {
				cssPrefix = fmt.Sprintf(".%s_h3_%d {\n",dObj.docName, dObj.h3Count)
				prefix = fmt.Sprintf("<h3 class=\"%s_h3_%d\"", dObj.docName, dObj.h3Count)
				dObj.h3Count++
			}
			suffix = "</h3>"
		case "HEADING_4":
			if dObj.h4.exist && !alter {
				prefix = fmt.Sprintf("<h4 class=\"%s_h4\"", dObj.docName)
				dObj.h4.exist = true
			}
			if alter {
				cssPrefix = fmt.Sprintf(".%s_h4_%d {\n",dObj.docName, dObj.h4Count)
				prefix = fmt.Sprintf("<h4 class=\"%s_h4_%d\"", dObj.docName, dObj.h4Count)
				dObj.h4Count++
			}
			suffix = "</h4>"
		case "HEADING_5":
			if dObj.h5.exist && !alter {
				prefix = fmt.Sprintf("<h5 class=\"%s_h5\"", dObj.docName)
				dObj.h5.exist = true
			}
			if alter {
				cssPrefix = fmt.Sprintf(".%s_h5_%d {\n",dObj.docName, dObj.h5Count)
				prefix = fmt.Sprintf("<h5 class=\"%s_h5_%d\"", dObj.docName, dObj.h5Count)
				dObj.h5Count++
			}
			suffix = "</h5>"
		case "HEADING_6":
			if dObj.h6.exist && !alter {
				prefix = fmt.Sprintf("<h6 class=\"%s_h6\"", dObj.docName)
				dObj.h6.exist = true
			}
			if alter {
				cssPrefix = fmt.Sprintf(".%s_h6_%d {\n",dObj.docName, dObj.h6Count)
				prefix = fmt.Sprintf("<h6 class=\"%s_h6_%d\"", dObj.docName, dObj.h6Count)
				dObj.h6Count++
			}
			suffix = "</h6>"
		case "NORMAL_TEXT":
			switch {
				case isList:
					prefix = "<span>"
					suffix = "</span>"
				case alter:
					cssPrefix = fmt.Sprintf(".%s_p_%d {\n",dObj.docName, dObj.parCount)
					prefix = fmt.Sprintf("<p class=\"%s_p_%d\"",dObj.docName, dObj.parCount)
					suffix = "\n</p>"
				default:
					prefix = fmt.Sprintf("<p class=\"%s_p\"", dObj.docName)
					suffix = "\n</p>"
			}
		case "NAMED_STYLE_TYPE_UNSPECIFIED":
//			namTypValid = false

		default:
//			namTypValid = false
	}

	if len(headingId) > 0 {
		prefix = fmt.Sprintf("%s id=\"%s\">", prefix, headingId[3:])
	} else {
		prefix = prefix + ">"
	}
//fmt.Printf("parstyl: %s %s %s\n", parStyl.NamedStyleType, prefix, suffix)
	if (len(cssPrefix) > 0) {cssStr = cssComment + cssPrefix + cssParAtt + "}\n"}

// test for a valid namestyl type
	return cssStr, prefix, suffix, nil
}


func (dObj *GdocHtmlObj) cvtTxtStylCss(txtStyl *docs.TextStyle, head bool)(cssStr string, err error) {
	var tcssStr string

	if txtStyl == nil {
		return "", fmt.Errorf("decode txtstyle: -- no Style")
	}

	if len(txtStyl.BaselineOffset) > 0 {
		switch txtStyl.BaselineOffset {
			case "SUPERSCRIPT":
				tcssStr += "  vertical-align: sub;\n"
			case "SUBSCRIPT":
				tcssStr += "	vertical-align: sup;\n"
			case "NONE":

			default:
				tcssStr += fmt.Sprintf("/* Baseline Offset unknown: %s */\n", txtStyl.BaselineOffset)
		}
	}
	if txtStyl.Bold {
		tcssStr += "  font-weight: bold;\n"
	} else {
		if head {tcssStr += "  font-weight: normal;\n"}
	}
	if txtStyl.Italic { tcssStr += "  font-style: italic;\n"}
	if txtStyl.Underline { tcssStr += "  text-decoration: underline;\n"}
	if txtStyl.Strikethrough { tcssStr += "  text-decoration: line-through;\n"}

	if txtStyl.WeightedFontFamily != nil {
		font := txtStyl.WeightedFontFamily.FontFamily
		weight := txtStyl.WeightedFontFamily.Weight
		tcssStr += fmt.Sprintf("  font-family: %s;\n", font)
		tcssStr += fmt.Sprintf("  font-weight: %d;\n", weight)
	}
	if txtStyl.FontSize != nil {
		mag := txtStyl.FontSize.Magnitude
		tcssStr += fmt.Sprintf("  font-size: %.2fpt;\n", mag)
	}
	if txtStyl.ForegroundColor != nil {
		if txtStyl.ForegroundColor.Color != nil {
			//0 to 1
            tcssStr += "  color: "
            tcssStr += getColor(txtStyl.ForegroundColor.Color)
		}
	}
	if txtStyl.BackgroundColor != nil {
		if txtStyl.BackgroundColor.Color != nil {
            tcssStr += "  background-color: "
            tcssStr += getColor(txtStyl.BackgroundColor.Color)
		}
	}

//	if txtStyl.Link != nil {
	//txtHtmlStr = txtHtmlStr + '<a href = "' + partAtts.LINK_URL + '">'

	if len(tcssStr) > 0 {
		cssStr = tcssStr
	}
	return cssStr, nil
}


func (dObj *GdocHtmlObj) createDivHead(divName string) (divObj *dispObj, err error) {
	var cssStr string
	//gdoc division css

	if len(divName) == 0 { return divObj, fmt.Errorf("createDivHead: no divNam!") }
	cssStr = fmt.Sprintf(".%s_%s {\n", dObj.docName, divName)

	// html
	divObj.bodyHtml = fmt.Sprintf("<div class=\"%s_div %s\">\n", dObj.docName, divName)
	// css
	divObj.bodyCss = cssStr

	return divObj, nil
}

func (dObj *GdocHtmlObj) createHead() (headObj dispObj, err error) {
	var cssStr string
	//gdoc division css
	cssStr = fmt.Sprintf(".%s_div {\n", dObj.docName)

    docstyl := dObj.doc.DocumentStyle
	if dObj.Options.Toc {
		cssStr += "  margin-top: 0mm;\n"
		cssStr += "  margin-bottom: 0mm;\n"
	} else {
		cssStr += fmt.Sprintf("  margin-top: %.1fmm; \n",docstyl.MarginTop.Magnitude*PtTomm)
		cssStr += fmt.Sprintf("  margin-bottom: %.1fmm; \n",docstyl.MarginBottom.Magnitude*PtTomm)
	}
    cssStr += fmt.Sprintf("  margin-right: %.2fmm; \n",docstyl.MarginRight.Magnitude*PtTomm)
    cssStr += fmt.Sprintf("  margin-left: %.2fmm; \n",docstyl.MarginLeft.Magnitude*PtTomm)

	dObj.docWidth = (docstyl.PageSize.Width.Magnitude - docstyl.MarginRight.Magnitude - docstyl.MarginLeft.Magnitude)*PtTomm

	cssStr += fmt.Sprintf("  width: %.1fmm;\n", dObj.docWidth)

	if dObj.Options.DivBorders {
		cssStr += "  border: solid red;\n"
		cssStr += "  border-width: 1px;\n"
	}

	// add default text style
	defTxtMap := new(textMap)
	parStyl, txtStyl, err := dObj.getNamedStyl("NORMAL_TEXT")
	if err != nil {
		return headObj, fmt.Errorf("creHeadCss: %v", err)
	}

	_, err = fillTxtMap(defTxtMap, txtStyl)
	if err != nil {
		return headObj, fmt.Errorf("creHeadCss: %v", err)
	}

	cssStr += cvtTxtMapCss(defTxtMap)
	cssStr += "}\n"

//plist
	// paragraph default style
	cssStr += fmt.Sprintf(".%s_p {\n", dObj.docName)
	cssStr += "  display: block;\n"

	defParMap := new(parMap)

	_, err = fillParMap(defParMap, parStyl)
	if err != nil {
//fmt.Printf("error %v\n", err)
		return headObj, fmt.Errorf("creHeadCss: %v", err)
	}
//fmt.Printf("txtmap: %v\n", defTxtMap)

	cssStr += dObj.cvtParMapCss(defParMap)

//	cssStr += txtCssStr

	cssStr += "}\n"

//	cssStr += fmt.Sprintf(".%s_p.list {display: inline;}\n", dObj.docName)

// list css strings

	for i:=0; i<len(dObj.docLists); i++ {
		listid := dObj.docLists[i].listId
		listClass := listid[4:]
		list := dObj.doc.Lists[listid]

		switch dObj.docLists[i].ord {
			case true:
				cssStr += fmt.Sprintf(".%s_ol {\n", listClass)

			case false:
				cssStr += fmt.Sprintf(".%s_ul {\n", listClass)
				cssStr += fmt.Sprintf("  list-style-type: none;\n")
				cssStr += fmt.Sprintf("  list-style-position: outside;\n")
				cssStr += fmt.Sprintf("}\n")
		}
		cssStr += fmt.Sprintf(".%s_li {\n", listClass)
		cssStr += fmt.Sprintf("  display: list-item;\n")
		cssStr += fmt.Sprintf("  text-align: start;\n")
		cssStr += fmt.Sprintf("  padding-left: 6pt;\n")
		cssStr += fmt.Sprintf("}\n")

		nestLev0 := list.ListProperties.NestingLevels[0]
		defGlyphTxtMap := new(textMap)
		_, err = fillTxtMap(defGlyphTxtMap, nestLev0.TextStyle)
		if err != nil { cssStr += "/* error def Glyph Text Style */\n" }

		for nl:=0; nl <= int(dObj.docLists[i].maxNestLev); nl++ {
			nestLev := list.ListProperties.NestingLevels[nl]
			glyphTxtMap := defGlyphTxtMap
			if nl > 0 {
				_, err := fillTxtMap(glyphTxtMap, nestLev.TextStyle)
				if err != nil { cssStr += "/* error def Glyph Text Style */\n" }
			}

			switch dObj.docLists[i].ord {
				case true:
					cssStr += fmt.Sprintf(".%s_ol.nL_%d {\n", listClass, nl)
				case false:
					cssStr += fmt.Sprintf(".%s_ul.nL_%d {\n", listClass, nl)
			}

			idFl := nestLev.IndentFirstLine.Magnitude
			idSt := nestLev.IndentStart.Magnitude
			cssStr += fmt.Sprintf("  margin: 0 0 0 %.0fpt;\n", idFl)
			cssStr += fmt.Sprintf("  padding-left: %.0fpt;\n", idSt-idFl - 6.0)
			cssStr += fmt.Sprintf("}\n")

			// Css <li nest level>
			cssStr += fmt.Sprintf(".%s_li.nL_%d {\n", listClass, nl)
			switch dObj.docLists[i].ord {
				case true:
//					cssStr += fmt.Sprintf("list-style-type: %s;\n", )
				case false:
					cssStr += dObj.cvtGlyph(nestLev)
			}
			cssStr += fmt.Sprintf("}\n")

			// Css marker
			cssStr += fmt.Sprintf(".%s_li.nL_%d::marker {\n", listClass, nl)
			cssStr +=  cvtTxtMapCss(glyphTxtMap)
			cssStr += fmt.Sprintf("}\n")
		}
	}

	if dObj.Options.Toc {
		cssStr += fmt.Sprintf(".%s_toc {\n", dObj.docName)
		cssStr += fmt.Sprintf("  margin-top: %.1fmm;\n", docstyl.MarginTop.Magnitude*PtTomm)
		cssStr += fmt.Sprintf("  margin-bottom: %.1fmm;\n}\n", docstyl.MarginBottom.Magnitude*PtTomm)
	}

	headObj.bodyCss = cssStr
	//gdoc division html
	headObj.bodyHtml = fmt.Sprintf("<div class=\"%s_div\">", dObj.docName)

	return headObj, nil
}

func (dObj *GdocHtmlObj) cvtContentEl(contEl *docs.StructuralElement) (GdocHtmlObj *dispObj, err error) {
	if dObj == nil {
		return nil, fmt.Errorf("error cvtContentEl: -- dObj is nil")
	}

	bodyElObj := new(dispObj)

	if contEl.Paragraph != nil {
		parEl := contEl.Paragraph
		tObj, err := dObj.cvtPar(parEl)
		if err != nil { bodyElObj.bodyHtml += fmt.Sprintf("<!-- %v -->\n", err) }
		addDispObj(bodyElObj, &tObj)
	}

	if contEl.SectionBreak != nil {

	}
	if contEl.Table != nil {
		tableEl := contEl.Table
		tObj, err := dObj.cvtTable(tableEl)
		if err != nil { bodyElObj.bodyHtml += fmt.Sprintf("<!-- %v -->\n", err) }
		addDispObj(bodyElObj, &tObj)
	}
	if contEl.TableOfContents != nil {

	}
//	fmt.Println(" ConvertEl: ",htmlObj)
	return bodyElObj, nil
}

func (dObj *GdocHtmlObj) cvtTocHeadCss() (CssStr string, err error) {
	var cssStr, tStr string
	var NamStyl *docs.NamedStyle

	if dObj == nil {
		return "", fmt.Errorf("/* error convertTocHeadtoCss -- dObj is nil */")
	}

	tocCssId := fmt.Sprintf("%s_TOC", dObj.docName)

	doc := dObj.doc
    docstyl := doc.DocumentStyle
	nStyl := doc.NamedStyles

	cssStr = "." + tocCssId + " {\n"
	cssStr += "  margin-top: 10mm;\n"
	cssStr += "  margin-bottom: 10mm;\n"
    cssStr += fmt.Sprintf("  margin-right: %.2fmm; \n",docstyl.MarginRight.Magnitude*PtTomm)
    cssStr += fmt.Sprintf("  margin-left: %.2fmm; \n",docstyl.MarginLeft.Magnitude*PtTomm)

	dObj.docWidth = (docstyl.PageSize.Width.Magnitude - docstyl.MarginRight.Magnitude - docstyl.MarginLeft.Magnitude)*PtTomm

	cssStr += fmt.Sprintf("  width: %.1fmm;\n", dObj.docWidth)
//xxx
	if dObj.Options.DivBorders {
		cssStr += "  border: solid green;\n"
		cssStr += "  border-width: 1px;\n"
	}
	cssStr += "  padding-top:10px;\n  padding-bottom:10px;\n"
	normStyl := -1
// find normal first

	for istyl:=0; istyl<len(nStyl.Styles); istyl++ {
		if nStyl.Styles[istyl].NamedStyleType == "NORMAL_TEXT" {
			normStyl = istyl
			NamStyl = nStyl.Styles[istyl]
			break
		}
	}

	if normStyl < 0 {
		return "", fmt.Errorf("ConvertNStyl -- no NORMAL_TEXT")
	}

	tStr, _, _, err = dObj.cvtParStyl(nil, NamStyl.ParagraphStyle, false)
	if err != nil {
		return cssStr, fmt.Errorf("cvtParStyl: %v",err)
	}
	cssStr += tStr
	tStr, err = dObj.cvtTxtStylCss(NamStyl.TextStyle, true)
	if err != nil {
		return cssStr, fmt.Errorf("cvtTxtStyl: %v",err)
	}

	cssStr += tStr + "}\n"

	for istyl:=0; istyl<len(nStyl.Styles); istyl++ {
		tStr = ""
		NamStyl := nStyl.Styles[istyl]
		switch NamStyl.NamedStyleType {
		case "TITLE":
			tStr =fmt.Sprintf("#%s_title {\n",tocCssId)
		case "SUBTITLE":
			tStr =fmt.Sprintf("#%s_subtitle {\n",tocCssId)
		case "HEADING_1":
			tStr =fmt.Sprintf(".%s h1 {\n",tocCssId)
 			tStr += "  padding-left: 10px;\n  margin: 0px;"
		case "HEADING_2":
			tStr =fmt.Sprintf(".%s h2 {\n",tocCssId)
			tStr += " padding-left: 20px;\n  margin: 0px;"
		case "HEADING_3":
			tStr =fmt.Sprintf(".%s h3 {\n", tocCssId)
			tStr += " padding-left: 40px;\n  margin: 0px;"
		case "HEADING_4":
			tStr =fmt.Sprintf(".%s h4 {\n", tocCssId)
			tStr += " padding-left: 60px;\n  margin: 0px;"
		case "HEADING_5":
			tStr =fmt.Sprintf(".%s h5 {\n", tocCssId)
			tStr += " padding-left: 80px;\n  margin: 0px;"
		case "HEADING_6":
			tStr =fmt.Sprintf(".%s h6 {\n", tocCssId)
			tStr += " padding-left: 100px;\n  margin: 0px;"
		case "NORMAL_TEXT":

		default:
			tStr =fmt.Sprintf("/* error - header: %s */", NamStyl.NamedStyleType)
		}


		cssStr += tStr
	}
	return cssStr, nil
}


func (dObj *GdocHtmlObj) creTocHead() (hdObj *dispObj, err error) {
	if dObj == nil {
		return nil, fmt.Errorf("creTocHead -- no GdocObj!")
	}

	hdObj = new(dispObj)
	hdObj.tocHtml = fmt.Sprintf("<div id=\"%s\"_TOC class=\"%s\">\n", dObj.docName, dObj.docName)
	hdObj.tocHtml += fmt.Sprintf("<p id=\"%s_TOC_subtitle\">Table of Contents</p>\n",dObj.docName)
	hdObj.tocCss, err = dObj.cvtTocHeadCss()
	if err != nil {
		return hdObj, fmt.Errorf("creTocHead:: cvtTocHeadCss: %v", err)
	}
	return hdObj, nil
}

func (dObj *GdocHtmlObj) cvtBody() (bodyObj *dispObj, err error) {

	if dObj == nil {
		return nil, fmt.Errorf("cvtBody -- no GdocObj!")
	}


	doc := dObj.doc
	body := doc.Body
	if body == nil {
		return nil, fmt.Errorf("cvtBody -- no body!")
	}
	bodyObj = new(dispObj)

//	toc := dObj.Options.Toc
	bodyObj.bodyHtml = fmt.Sprintf("<div class=\"%s_div\">\n", dObj.docName)

	elNum := len(body.Content)
	for el:=0; el< elNum; el++ {
		bodyEl := body.Content[el]
		tObj, err:=dObj.cvtContentEl(bodyEl)
		if err != nil {
			fmt.Println("cvtContentEl: %v", err)
		}
//		fmt.Println("tObj:", tObj)
		addDispObj(bodyObj, tObj)
	} // for el loop end
	if dObj.listStack != nil {
		bodyObj.bodyHtml += dObj.closeList(len(*dObj.listStack))
//fmt.Printf("end of doc closing list!")
	}

	bodyObj.tocHtml += "</div>\n\n"
	bodyObj.bodyHtml += "</div>\n\n"

	return bodyObj, nil
}

func (dObj *GdocHtmlObj) cvtBodySec(elSt, elEnd int) (bodyObj *dispObj, err error) {

	if dObj == nil {
		return nil, fmt.Errorf("cvtBody -- no GdocObj!")
	}


	doc := dObj.doc
	body := doc.Body
	if body == nil {
		return nil, fmt.Errorf("cvtBody -- no body!")
	}

	elCount := len(body.Content)

	if elSt > elCount { return nil, fmt.Errorf("cvtSelBody: elSt > elCount!") }
	if elEnd > elCount { return nil, fmt.Errorf("cvtSelBody: elEnd > elCount!") }
	if elSt > elEnd { return nil, fmt.Errorf("cvtSelBody: elSt > elElnd!") }

//	toc := dObj.Options.Toc
	bodyObj = new(dispObj)

	// need to move
	bodyObj.bodyHtml = fmt.Sprintf("<div class=\"%s_div\">\n", dObj.docName)

	for el:=0; el< elCount; el++ {
		bodyEl := body.Content[el]
		tObj, err:=dObj.cvtContentEl(bodyEl)
		if err != nil {
			fmt.Println("cvtContentEl: %v", err)
		}
		addDispObj(bodyObj, tObj)
	} // for el loop end

	if dObj.listStack != nil {
		bodyObj.bodyHtml += dObj.closeList(len(*dObj.listStack))
//fmt.Printf("end of doc closing list!")
	}

	bodyObj.tocHtml += "</div>\n\n"
	bodyObj.bodyHtml += "</div>\n\n"

	return bodyObj, nil
}

func CreGdocHtmlDoc(folderPath string, doc *docs.Document, options *OptObj)(err error) {
	// function which converts the entire document into an hmlt file
	var tocDiv *dispObj
	var dObj GdocHtmlObj

	err = dObj.initGdocHtml(doc, options)
	if err != nil {
		return fmt.Errorf("initGdocHtml %v", err)
	}

	err = dObj.createHtmlFolder(folderPath)
	if err!= nil {
		return fmt.Errorf("createHtmlFolder %v", err)
	}

	err = dObj.createOutFil("")
	if err!= nil {
		return fmt.Errorf("createOutFil %v", err)
	}

	if dObj.Options.ImgFold {
		err = dObj.dlImages()
		if err != nil {
			fmt.Errorf("dlImages: %v", err)
		}
	}

	mainDiv, err := dObj.cvtBody()
	if err != nil {
		return fmt.Errorf("cvtBody: %v", err)
	}

	headObj, err := dObj.createHead()
	if err != nil {
		return fmt.Errorf("creHeadCss: %v", err)
	}

	toc := dObj.Options.Toc
	if toc {
		tocDiv, err = dObj.creTocHead()
		if err != nil {
			tocDiv.tocHtml = fmt.Sprintf("<!--- error Toc Head: %v --->\n",err)
		}
	}

	// create html file
	outfil := dObj.htmlFil
	docHeadStr,_ := creHtmlHead()
	outfil.WriteString(docHeadStr)

	// div Css and named styles used
	outfil.WriteString(headObj.bodyCss)

	outfil.WriteString(mainDiv.headCss)
	outfil.WriteString(mainDiv.bodyCss)
	if toc {
		outfil.WriteString(tocDiv.tocCss)
		outfil.WriteString(mainDiv.tocCss)
	}
	outfil.WriteString("</style>\n</head>\n<body>\n")

	// init html comments
	outfil.WriteString(headObj.bodyHtml)

	if toc {outfil.WriteString(tocDiv.tocHtml)}

	outfil.WriteString(mainDiv.bodyHtml)
	outfil.WriteString("</body>\n</html>\n")
	outfil.Close()
	return nil
}

func CreGdocHtmlMain(folderPath string, doc *docs.Document, options *OptObj)(err error) {
// function that converts the main part of a gdoc document into an html file
// excludes everything before the "main" heading or
// excludes sections titled "summary" and "keywords"

	var tocDiv *dispObj
	var dObj GdocHtmlObj

	err = dObj.initGdocHtml(doc, options)
	if err != nil {
		return fmt.Errorf("initGdocHtml: %v", err)
	}

	err = dObj.createHtmlFolder(folderPath)
	if err!= nil {
		return fmt.Errorf("createHtmlFolder: %v", err)
	}

	err = dObj.createOutFil("main")
	if err!= nil {
		return fmt.Errorf("createOutFil: %v", err)
	}

	if dObj.Options.ImgFold {
		err = dObj.dlImages()
		if err != nil {
			fmt.Errorf("dlImages: %v", err)
		}
	}

	mainDiv, err := dObj.cvtBody()
	if err != nil {
		return fmt.Errorf("cvtBody: %v", err)
	}

	headObj, err := dObj.createHead()
	if err != nil {
		return fmt.Errorf("creHeadCss: %v", err)
	}

	toc := dObj.Options.Toc
	if toc {
		tocDiv, err = dObj.creTocHead()
		if err != nil {
			tocDiv.tocHtml = fmt.Sprintf("<!--- error Toc Head: %v --->\n",err)
		}
	}

	// create html file
	outfil := dObj.htmlFil
	if outfil == nil {return fmt.Errorf("outfil is nil!")}
	docHeadStr,_ := creHtmlHead()
	outfil.WriteString(docHeadStr)
	// basic Css
	outfil.WriteString(headObj.bodyCss)
	// named styles
	outfil.WriteString(mainDiv.headCss)
	outfil.WriteString(mainDiv.bodyCss)
	if toc {
		outfil.WriteString(tocDiv.tocCss)
		outfil.WriteString(mainDiv.tocCss)
	}
	outfil.WriteString("</style>\n</head>\n<body>\n")

	outfil.WriteString(headObj.bodyHtml)
	if toc {outfil.WriteString(tocDiv.tocHtml)}

	outfil.WriteString(mainDiv.bodyHtml)
	outfil.WriteString("</body>\n</html>\n")
	outfil.Close()
	return nil
}

func CreGdocHtmlSection(heading, folderPath string, doc *docs.Document, options *OptObj)(err error) {
// function that creates an html fil from the named section
	var tocDiv *dispObj
	var dObj GdocHtmlObj

	if len(heading) == 0 {fmt.Errorf("CreGdocHtmlSection:: no heading string provided!")}

	err = dObj.initGdocHtml(doc, options)
	if err != nil {
		return fmt.Errorf("initGdocHtml: %v", err)

	}

	err = dObj.createHtmlFolder(folderPath)
	if err!= nil {
		return fmt.Errorf("createHtmlFolder: %v", err)
	}

	err = dObj.createOutFil("heading")
	if err!= nil {
		return fmt.Errorf("createOutFil: %v", err)
	}

	if dObj.Options.ImgFold {
		err = dObj.dlImages()
		if err != nil {
			fmt.Errorf("dlImages: %v", err)
		}
	}

	mainDiv, err := dObj.cvtBody()
	if err != nil {
		return fmt.Errorf("cvtBody: %v", err)
	}

	headObj, err := dObj.createHead()
	if err != nil {
		return fmt.Errorf("creHeadCss: %v", err)
	}

	toc := dObj.Options.Toc
	if toc {
		tocDiv, err = dObj.creTocHead()
		if err != nil {
			tocDiv.tocHtml = fmt.Sprintf("<!--- error Toc Head: %v --->\n",err)
		}
	}

	// create html file
	outfil := dObj.htmlFil
	if outfil == nil {return fmt.Errorf("outfil is nil!")}

	docHeadStr,_ := creHtmlHead()
	outfil.WriteString(docHeadStr)
	// basic Css
	outfil.WriteString(headObj.bodyCss)
	// named styles
	outfil.WriteString(mainDiv.headCss)
	outfil.WriteString(mainDiv.bodyCss)
	if toc {
		outfil.WriteString(tocDiv.tocCss)
		outfil.WriteString(mainDiv.tocCss)
	}
	outfil.WriteString("</style>\n</head>\n<body>\n")

	outfil.WriteString(headObj.bodyHtml)
	if toc {outfil.WriteString(tocDiv.tocHtml)}
	outfil.WriteString(mainDiv.bodyHtml)

	outfil.WriteString("</body>\n</html>\n")
	outfil.Close()
	return nil
}

func CreGdocHtmlAll(folderPath string, doc *docs.Document, options *OptObj)(err error) {
// function that creates an html fil from the named section
	var tocDiv *dispObj
	var dObj GdocHtmlObj

	err = dObj.initGdocHtml(doc, options)
	if err != nil {
		return fmt.Errorf("initGdocHtml %v", err)
	}

	err = dObj.createHtmlFolder(folderPath)
	if err!= nil {
		return fmt.Errorf("createHtmlFolder: %v", err)
	}

	err = dObj.createOutFil("")
	if err!= nil {
		return fmt.Errorf("createOutFil %v", err)
	}

	if dObj.Options.ImgFold {
		err = dObj.dlImages()
		if err != nil {
			fmt.Errorf("dlImages: %v", err)
		}
	}

	mainDiv, err := dObj.cvtBody()
	if err != nil {
		return fmt.Errorf("cvtBody: %v", err)
	}

	headObj, err := dObj.createHead()
	if err != nil {
		return fmt.Errorf("createHead: %v", err)
	}

	toc := dObj.Options.Toc
	if toc {
		tocDiv, err = dObj.creTocHead()
		if err != nil {
			tocDiv.tocHtml = fmt.Sprintf("<!--- error Toc Head: %v --->\n",err)
		}
	}

	// create html file
	outfil := dObj.htmlFil
	if outfil == nil {
		return fmt.Errorf("outfil is nil!")
	}
	docHeadStr,_ := creHtmlHead()
	outfil.WriteString(docHeadStr)
	// basic Css
	outfil.WriteString(headObj.bodyCss)
	// named styles
	outfil.WriteString(mainDiv.headCss)
	outfil.WriteString(mainDiv.bodyCss)
	if toc {
		outfil.WriteString(tocDiv.tocCss)
		outfil.WriteString(mainDiv.tocCss)
	}
	outfil.WriteString("</style>\n</head>\n<body>\n")

	outfil.WriteString(headObj.bodyHtml)
	if toc {outfil.WriteString(tocDiv.tocHtml)}
	outfil.WriteString(mainDiv.bodyHtml)

	outfil.WriteString("</body>\n</html>\n")
	outfil.Close()
	return nil
}

