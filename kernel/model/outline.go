// SiYuan - Refactor your thinking
// Copyright (c) 2020-present, b3log.org
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package model

import (
	"github.com/88250/gulu"
	"github.com/88250/lute/ast"
	"github.com/88250/lute/html"
	"github.com/88250/lute/parse"
	"github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/siyuan-note/siyuan/kernel/av"
	"github.com/siyuan-note/siyuan/kernel/treenode"
	"github.com/siyuan-note/siyuan/kernel/util"
	"strconv"
	"strings"
	"time"
)

func (tx *Transaction) doMoveOutlineHeading(operation *Operation) (ret *TxErr) {
	headingID := operation.ID
	previousID := operation.PreviousID
	parentID := operation.ParentID

	tree, err := tx.loadTree(headingID)
	if err != nil {
		return &TxErr{code: TxErrCodeBlockNotFound, id: headingID}
	}
	operation.RetData = tree.Root.ID

	if headingID == parentID || headingID == previousID {
		return
	}

	heading := treenode.GetNodeInTree(tree, headingID)
	if nil == heading {
		return &TxErr{code: TxErrCodeBlockNotFound, id: headingID}
	}

	if ast.NodeDocument != heading.Parent.Type {
		// 仅支持文档根节点下第一层标题，不支持容器块内标题
		util.PushMsg(Conf.language(240), 5000)
		return
	}

	headings := []*ast.Node{}
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering && ast.NodeHeading == n.Type && !n.ParentIs(ast.NodeBlockquote) {
			headings = append(headings, n)
		}
		return ast.WalkContinue
	})

	headingChildren := treenode.HeadingChildren(heading)

	if "" != previousID {
		previousHeading := treenode.GetNodeInTree(tree, previousID)
		if nil == previousHeading {
			return &TxErr{code: TxErrCodeBlockNotFound, id: previousID}
		}

		if ast.NodeDocument != previousHeading.Parent.Type {
			// 仅支持文档根节点下第一层标题，不支持容器块内标题
			util.PushMsg(Conf.language(248), 5000)
			return
		}

		for _, h := range headingChildren {
			if h.ID == previousID {
				// 不能移动到自己的子标题下
				util.PushMsg(Conf.language(241), 5000)
				return
			}
		}

		generateOpTypeHistory(tree, HistoryOpOutline)

		targetNode := previousHeading
		previousHeadingChildren := treenode.HeadingChildren(previousHeading)
		if 0 < len(previousHeadingChildren) {
			targetNode = previousHeadingChildren[len(previousHeadingChildren)-1]
		}

		for _, h := range headingChildren {
			if h.ID == targetNode.ID { // 目标节点是当前标题的子节点
				targetNode = heading.Previous
			}
		}
		if targetNode.ID == heading.ID {
			targetNode = heading.Previous
		}

		diffLevel := heading.HeadingLevel - previousHeading.HeadingLevel
		heading.HeadingLevel = previousHeading.HeadingLevel

		for i := len(headingChildren) - 1; i >= 0; i-- {
			child := headingChildren[i]
			if ast.NodeHeading == child.Type {
				child.HeadingLevel -= diffLevel
				if 6 < child.HeadingLevel {
					child.HeadingLevel = 6
				}
			}
			targetNode.InsertAfter(child)
		}
		targetNode.InsertAfter(heading)
	} else if "" != parentID {
		parentHeading := treenode.GetNodeInTree(tree, parentID)
		if nil == parentHeading {
			return &TxErr{code: TxErrCodeBlockNotFound, id: parentID}
		}

		if ast.NodeDocument != parentHeading.Parent.Type {
			// 仅支持文档根节点下第一层标题，不支持容器块内标题
			util.PushMsg(Conf.language(248), 5000)
			return
		}

		for _, h := range headingChildren {
			if h.ID == parentID {
				// 不能移动到自己的子标题下
				util.PushMsg(Conf.language(241), 5000)
				return
			}
		}

		generateOpTypeHistory(tree, HistoryOpOutline)

		targetNode := parentHeading
		parentHeadingChildren := treenode.HeadingChildren(parentHeading)
		// 找到下方第一个非标题节点
		var tmp []*ast.Node
		for _, child := range parentHeadingChildren {
			if ast.NodeHeading == child.Type {
				break
			}
			tmp = append(tmp, child)
		}
		parentHeadingChildren = tmp
		if 0 < len(parentHeadingChildren) {
			for _, child := range parentHeadingChildren {
				if child.ID == headingID {
					break
				}
				targetNode = child
			}
		}

		diffLevel := heading.HeadingLevel - parentHeading.HeadingLevel - 1
		heading.HeadingLevel = parentHeading.HeadingLevel + 1
		if 6 < heading.HeadingLevel {
			heading.HeadingLevel = 6
		}

		for i := len(headingChildren) - 1; i >= 0; i-- {
			child := headingChildren[i]
			if ast.NodeHeading == child.Type {
				child.HeadingLevel -= diffLevel
				if 6 < child.HeadingLevel {
					child.HeadingLevel = 6
				}
			}
			targetNode.InsertAfter(child)
		}
		targetNode.InsertAfter(heading)
	} else {
		generateOpTypeHistory(tree, HistoryOpOutline)

		// 移到第一个标题前
		var firstHeading *ast.Node
		for n := tree.Root.FirstChild; nil != n; n = n.Next {
			if ast.NodeHeading == n.Type {
				firstHeading = n
				break
			}
		}
		if nil == firstHeading || firstHeading.ID == heading.ID {
			return
		}

		diffLevel := heading.HeadingLevel - firstHeading.HeadingLevel
		heading.HeadingLevel = firstHeading.HeadingLevel

		firstHeading.InsertBefore(heading)
		for i := 0; i < len(headingChildren); i++ {
			child := headingChildren[i]
			if ast.NodeHeading == child.Type {
				child.HeadingLevel -= diffLevel
				if 6 < child.HeadingLevel {
					child.HeadingLevel = 6
				}
			}
			firstHeading.InsertBefore(child)
		}
	}

	if err = tx.writeTree(tree); err != nil {
		return
	}
	return
}

func Outline(rootID string, preview bool) (ret []*Path, err error) {
	FlushTxQueue()

	ret = []*Path{}
	tree, _ := LoadTreeByBlockID(rootID)
	if nil == tree {
		return
	}

	if preview && Conf.Export.AddTitle {
		if root, _ := getBlock(tree.ID, tree); nil != root {
			root.IAL["type"] = "doc"
			title := &ast.Node{ID: root.ID, Type: ast.NodeHeading, HeadingLevel: 1}
			for k, v := range root.IAL {
				if "type" == k {
					continue
				}
				title.SetIALAttr(k, v)
			}
			title.InsertAfter(&ast.Node{Type: ast.NodeKramdownBlockIAL, Tokens: parse.IAL2Tokens(title.KramdownIAL)})

			content := html.UnescapeString(root.Content)
			title.AppendChild(&ast.Node{Type: ast.NodeText, Tokens: []byte(content)})
			tree.Root.PrependChild(title)
		}
	}

	ret = outline(tree)

	storage, _ := GetOutlineStorage(rootID)
	if nil == storage || 0 == len(storage) {
		// 默认全部展开
		for _, p := range ret {
			p.Folded = false
			for _, b := range p.Blocks {
				b.Folded = false
				for _, c := range b.Children {
					walkChildren(c, []string{"expandAll"})
				}
			}
		}
	}

	if nil != storage["expandIds"] {
		// 先全部折叠，后面再根据展开 ID 列表展开对应标题
		for _, p := range ret {
			p.Folded = true
			for _, b := range p.Blocks {
				b.Folded = true
				for _, c := range b.Children {
					walkChildren(c, []string{"expandNone"})
				}
			}
		}

		expandIDsArg := storage["expandIds"].([]interface{})
		var expandIDs []string
		for _, id := range expandIDsArg {
			expandIDs = append(expandIDs, id.(string))
		}

		for _, p := range ret {
			p.Folded = !gulu.Str.Contains(p.ID, expandIDs)
			for _, b := range p.Blocks {
				b.Folded = !gulu.Str.Contains(b.ID, expandIDs)
				for _, c := range b.Children {
					walkChildren(c, expandIDs)
				}
			}
		}
	}
	return
}

func walkChildren(b *Block, expandIDs []string) {
	if 1 == len(expandIDs) {
		if "expandAll" == expandIDs[0] {
			b.Folded = false
		} else if "expandNone" == expandIDs[0] {
			b.Folded = true
		} else {
			b.Folded = !gulu.Str.Contains(b.ID, expandIDs)
		}
	} else {
		b.Folded = !gulu.Str.Contains(b.ID, expandIDs)
	}

	for _, c := range b.Children {
		walkChildren(c, expandIDs)
	}
}

func outline(tree *parse.Tree) (ret []*Path) {
	luteEngine := NewLute()
	var headings []*Block
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if entering && ast.NodeHeading == n.Type && !n.ParentIs(ast.NodeBlockquote) {
			n.Box, n.Path = tree.Box, tree.Path
			block := &Block{
				RootID:  tree.Root.ID,
				Depth:   n.HeadingLevel,
				Box:     n.Box,
				Path:    n.Path,
				ID:      n.ID,
				Content: renderOutline(n, luteEngine),
				Type:    n.Type.String(),
				SubType: treenode.SubTypeAbbr(n),
				Folded:  true,
			}
			headings = append(headings, block)
			return ast.WalkSkipChildren
		}
		return ast.WalkContinue
	})

	if 1 > len(headings) {
		return
	}

	var blocks []*Block
	stack := linkedliststack.New()
	for _, h := range headings {
	L:
		for ; ; stack.Pop() {
			cur, ok := stack.Peek()
			if !ok {
				blocks = append(blocks, h)
				stack.Push(h)
				break L
			}

			tip := cur.(*Block)
			if tip.Depth < h.Depth {
				tip.Children = append(tip.Children, h)
				stack.Push(h)
				break L
			}
			tip.Count = len(tip.Children)
		}
	}

	ret = toFlatTree(blocks, 0, "outline", tree)
	if 0 < len(ret) {
		children := ret[0].Blocks
		ret = nil
		for _, b := range children {
			resetDepth(b, 0)
			ret = append(ret, &Path{
				ID:       b.ID,
				Box:      b.Box,
				Name:     b.Content,
				NodeType: b.Type,
				Type:     "outline",
				SubType:  b.SubType,
				Blocks:   b.Children,
				Depth:    0,
				Count:    b.Count,
				Folded:   true,
			})
		}
	}
	return
}

func resetDepth(b *Block, depth int) {
	b.Depth = depth
	b.Count = len(b.Children)
	for _, c := range b.Children {
		resetDepth(c, depth+1)
	}
}

func CollectTimestampedAIBlocks(rootID string, preview bool) (ret []*Path, err error) {
	FlushTxQueue()

	ret = []*Path{}
	tree, _ := LoadTreeByBlockID(rootID)
	if nil == tree {
		return
	}

	ret = collectTimestampedAI(tree)
	return
}

func collectTimestampedAI(tree *parse.Tree) (ret []*Path) {
	// Сначала собираем блоки с AV и алиасами
	avAliasBlocks := collectAVAndAliasBlocks(tree)
	if len(avAliasBlocks) > 0 && avAliasBlocks[0].Count > 0 {
		ret = append(ret, avAliasBlocks...)
	}

	// Сбор блоков с timestamp
	timestampedBlocks := collectTimestampedBlocks(tree)

	// Группировка по датам
	timestampedPaths := groupTimestampedBlocksByDate(timestampedBlocks, 1)

	// Оборачиваем хронологические блоки в отдельную папку
	if len(timestampedPaths) > 0 {
		chronoFolder := &Path{
			ID:       "chrono-folder",
			Name:     "Chronological",
			NodeType: "Folder",
			SubType:  "",
			Children: timestampedPaths,
			Depth:    0,
			Count:    len(timestampedPaths),
		}
		ret = append(ret, chronoFolder)
	}

	return
}

func collectAVAndAliasBlocks(tree *parse.Tree) []*Path {
	luteEngine := NewLute()

	// Карта для группировки блоков по AV
	avGroups := make(map[string][]*Block)
	var aliasOnlyBlocks []*Block

	// Проходим по всем блокам дерева
	ast.Walk(tree.Root, func(n *ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.WalkContinue
		}

		// Проверяем наличие custom-avs или alias
		avs := n.IALAttr("custom-avs")
		alias := n.IALAttr("alias")

		if "" == avs && "" == alias {
			return ast.WalkContinue
		}

		// Определяем содержимое для отображения
		displayContent := ""
		if "" != alias {
			displayContent = alias
		}

		// Если нет алиаса, берем краткое содержимое блока
		if "" == displayContent {
			blockContent := renderOutline(n, luteEngine)
			if "" != blockContent && len(blockContent) > 50 {
				blockContent = blockContent[:50] + "..."
			}
			displayContent = blockContent
		}

		block := &Block{
			ID:      n.ID,
			Box:     tree.Box,
			Content: displayContent,
			Type:    n.Type.String(),
			SubType: treenode.SubTypeAbbr(n),
			Depth:   1,
		}

		if "" != avs {
			// Разбираем custom-avs по запятым и добавляем блок в каждую группу
			avIDs := strings.Split(avs, ",")
			for _, avID := range avIDs {
				avID = strings.TrimSpace(avID)
				if "" == avID {
					continue
				}

				// Получаем имя AV
				avName, err := av.GetAttributeViewName(avID)
				if nil != err {
					avName = avID // Используем ID если не удалось получить имя
				}
				if "" == avName {
					avName = Conf.language(105) // "Untitled"
				}

				// Добавляем блок в группу для этого AV
				avGroups[avName] = append(avGroups[avName], block)
			}
		} else {
			// Блок только с алиасом
			aliasOnlyBlocks = append(aliasOnlyBlocks, block)
		}

		return ast.WalkContinue
	})

	var result []*Path

	// Создаем Path для каждого AV
	for avName, blocks := range avGroups {
		avPath := &Path{
			ID:     blocks[0].ID, // ID первого блока в группе
			Box:    blocks[0].Box,
			Name:   avName,
			Type:   "av-group",
			Blocks: blocks,
			Count:  len(blocks),
			Depth:  0,
		}
		result = append(result, avPath)
	}

	// Создаем отдельный Path для алиасов без AV
	if len(aliasOnlyBlocks) > 0 {
		aliasPath := &Path{
			ID:     aliasOnlyBlocks[0].ID,
			Box:    aliasOnlyBlocks[0].Box,
			Name:   "Aliases",
			Type:   "alias-group",
			Blocks: aliasOnlyBlocks,
			Count:  len(aliasOnlyBlocks),
			Depth:  0,
		}
		result = append(result, aliasPath)
	}

	// Если ничего не нашли, возвращаем пустой результат
	if len(result) == 0 {
		return []*Path{{ID: "av-alias-table", Name: "AV & Aliases", Type: "av-alias", Blocks: []*Block{}, Count: 0}}
	}

	return result
}

// Структура для хранения Path с метаданными времени
type timestampedPath struct {
	path      *Path
	timestamp int64
	remainder string
}

func collectTimestampedBlocks(tree *parse.Tree) []*timestampedPath {
	luteEngine := NewLute()
	var timestampedPaths []*timestampedPath

	// Проходим по прямым детям root
	for n := tree.Root.FirstChild; nil != n; n = n.Next {
		if ast.NodeSuperBlock != n.Type {
			continue
		}

		// Проверяем наличие атрибута custom-timestamp
		timestampStr := n.IALAttr("custom-timestamp")
		if "" == timestampStr {
			continue
		}

		// Получаем remainder если есть
		remainder := n.IALAttr("custom-timestamp-remainder")

		// Парсим timestamp
		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			continue
		}

		blockTime := time.Unix(timestamp, 0)
		formattedTime := blockTime.Format("02-Jan-06 15:04")
		if "" != remainder {
			formattedTime = formattedTime + " <strong><i>" + remainder + "</i></strong>"
		}

		// Создаём Path для SuperBlock
		superPath := &Path{
			ID:       n.ID,
			Box:      tree.Box,
			Name:     formattedTime,
			NodeType: n.Type.String(),
			Type:     "timestamp-ai",
			SubType:  treenode.SubTypeAbbr(n),
			Blocks:   []*Block{}, // Используем Blocks для хранения дочерних блоков
		}

		// Ищем вложенные blockquote
		ast.Walk(n, func(child *ast.Node, entering bool) ast.WalkStatus {
			if !entering {
				return ast.WalkContinue
			}

			if ast.NodeBlockquote != child.Type {
				return ast.WalkContinue
			}

			aiBlock := &Block{
				ID:      child.ID,
				Box:     tree.Box,
				Content: renderOutline(child, luteEngine),
				Type:    child.Type.String(),
				SubType: treenode.SubTypeAbbr(child),
				Depth:   1,
			}
			superPath.Blocks = append(superPath.Blocks, aiBlock)

			return ast.WalkContinue
		})

		superPath.Count = len(superPath.Blocks)

		// Добавляем в список с метаданными
		timestampedPaths = append(timestampedPaths, &timestampedPath{
			path:      superPath,
			timestamp: timestamp,
			remainder: remainder,
		})
	}

	return timestampedPaths
}

func groupTimestampedBlocksByDate(timestampedBlocks []*timestampedPath, initialDepth int) []*Path {
	var ret []*Path

	// Группируем блоки по датам используя time.Time
	type dateGroup struct {
		date   time.Time
		blocks []*timestampedPath
	}
	dateGroups := make(map[string]*dateGroup)
	var dates []string // для сохранения порядка

	for _, block := range timestampedBlocks {
		blockTime := time.Unix(block.timestamp, 0)
		// Получаем дату без времени (начало дня)
		dateOnly := time.Date(blockTime.Year(), blockTime.Month(), blockTime.Day(), 0, 0, 0, 0, blockTime.Location())
		dateKey := dateOnly.Format("2006-01-02")

		if _, exists := dateGroups[dateKey]; !exists {
			dates = append(dates, dateKey)
			dateGroups[dateKey] = &dateGroup{
				date:   dateOnly,
				blocks: []*timestampedPath{},
			}
		}
		dateGroups[dateKey].blocks = append(dateGroups[dateKey].blocks, block)
	}

	// Преобразуем в формат Path с группировкой
	for _, dateKey := range dates {
		group := dateGroups[dateKey]
		blocks := group.blocks

		if len(blocks) > 1 {
			// Создаем родительский Path для группы с одинаковой датой
			groupPath := &Path{
				ID:       blocks[0].path.ID, // ID первого блока в группе
				Box:      blocks[0].path.Box,
				Name:     group.date.Format("02-Jan-06"), // Форматируем дату для отображения
				NodeType: "Group",
				Type:     "timestamp-ai",
				SubType:  "",
				Blocks:   []*Block{}, // Используем Blocks вместо Children
				Depth:    initialDepth,
				Count:    0,
			}

			// Добавляем первым дочерним элементом дубликат с временем первого блока
			firstBlockTime := time.Unix(blocks[0].timestamp, 0)
			firstTimeDisplay := firstBlockTime.Format("15:04")
			if "" != blocks[0].remainder {
				firstTimeDisplay = firstTimeDisplay + " <strong><i>" + blocks[0].remainder + "</i></strong>"
			}

			firstTimeBlock := &Block{
				ID:       blocks[0].path.ID,
				Box:      blocks[0].path.Box,
				Content:  firstTimeDisplay,
				Type:     blocks[0].path.NodeType,
				SubType:  blocks[0].path.SubType,
				Children: adjustBlockChildrenDepth(blocks[0].path.Blocks, initialDepth+2),
				Depth:    initialDepth + 1,
				Count:    blocks[0].path.Count,
			}
			groupPath.Blocks = append(groupPath.Blocks, firstTimeBlock)
			groupPath.Count++

			// Добавляем остальные блоки со временем (начиная со второго)
			for i := 1; i < len(blocks); i++ {
				block := blocks[i]
				blockTime := time.Unix(block.timestamp, 0)
				timeDisplay := blockTime.Format("15:04")
				if "" != block.remainder {
					timeDisplay = timeDisplay + " <strong><i>" + block.remainder + "</i></strong>"
				}

				timeBlock := &Block{
					ID:       block.path.ID,
					Box:      block.path.Box,
					Content:  timeDisplay,
					Type:     block.path.NodeType,
					SubType:  block.path.SubType,
					Children: adjustBlockChildrenDepth(block.path.Blocks, initialDepth+2),
					Depth:    initialDepth + 1,
					Count:    block.path.Count,
				}
				groupPath.Blocks = append(groupPath.Blocks, timeBlock)
				groupPath.Count++
			}

			ret = append(ret, groupPath)
		} else if len(blocks) == 1 {
			// Одиночный блок без группировки - показываем полный timestamp
			ret = append(ret, blocks[0].path)
		}
	}

	return ret
}

// Рекурсивно увеличивает глубину всех дочерних блоков
func adjustBlockChildrenDepth(children []*Block, newDepth int) []*Block {
	if len(children) == 0 {
		return nil
	}
	adjusted := make([]*Block, len(children))
	for i, child := range children {
		adjusted[i] = &Block{
			ID:       child.ID,
			Box:      child.Box,
			Content:  child.Content,
			Type:     child.Type,
			SubType:  child.SubType,
			Children: adjustBlockChildrenDepth(child.Children, newDepth+1),
			Depth:    newDepth,
			Count:    child.Count,
		}
	}
	return adjusted
}
