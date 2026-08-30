package watchlist

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	s "github.com/achannarasappa/ticker/v5/internal/sorter"
	row "github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist/row"
	u "github.com/achannarasappa/ticker/v5/internal/ui/util"

	grid "github.com/achannarasappa/term-grid"
	tea "github.com/charmbracelet/bubbletea"
)

// Config represents the configuration for the watchlist component
type Config struct {
	Separate              bool
	ShowPositions         bool
	ExtraInfoExchange     bool
	ExtraInfoFundamentals bool
	ShowOHLC              bool
	Sort                  string
	Styles                c.Styles
}

// Model for watchlist section
type Model struct {
	width          int
	assets         []*c.Asset
	assetsBySymbol map[string]*c.Asset
	sorter         s.Sorter
	config         Config
	cellWidths     row.CellWidthsContainer
	rows           []*row.Model
	rowsBySymbol   map[string]*row.Model
}

// Messages for replacing assets
type SetAssetsMsg []c.Asset

// Messages for updating assets
type UpdateAssetsMsg []c.Asset

// Messages for changing sort
type ChangeSortMsg string

// NewModel returns a model with default values
func NewModel(config Config) *Model {
	return &Model{
		width:          80,
		config:         config,
		assets:         make([]*c.Asset, 0),
		assetsBySymbol: make(map[string]*c.Asset),
		sorter:         s.NewSorter(config.Sort),
		rowsBySymbol:   make(map[string]*row.Model),
	}
}

// Init initializes the watchlist
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages for the watchlist
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SetAssetsMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// Convert []c.Asset to []*c.Asset and update assetsBySymbol map
		assets := make([]*c.Asset, len(msg))
		assetsBySymbol := make(map[string]*c.Asset)

		for i := range msg {
			assets[i] = &msg[i]
			assetsBySymbol[msg[i].Symbol] = assets[i]
		}

		assets = m.sorter(assets)

		for i, asset := range assets {
			if i < len(m.rows) {
				m.rows[i], cmd = m.rows[i].Update(row.UpdateAssetMsg(asset))
				cmds = append(cmds, cmd)
				m.rowsBySymbol[assets[i].Symbol] = m.rows[i]
			} else {
				m.rows = append(m.rows, row.New(row.Config{
					Separate:              m.config.Separate,
					ExtraInfoExchange:     m.config.ExtraInfoExchange,
					ExtraInfoFundamentals: m.config.ExtraInfoFundamentals,
					ShowOHLC:              m.config.ShowOHLC,
					ShowPositions:         m.config.ShowPositions,
					Styles:                m.config.Styles,
					Asset:                 asset,
				}))
				m.rowsBySymbol[assets[i].Symbol] = m.rows[len(m.rows)-1]
			}
		}

		if len(assets) < len(m.rows) {
			m.rows = m.rows[:len(assets)]
		}

		m.assets = assets
		m.assetsBySymbol = assetsBySymbol

		// TODO: only set conditionally if all assets have changed
		m.cellWidths = getCellWidths(m.assets)
		for i, r := range m.rows {
			m.rows[i], _ = r.Update(row.SetCellWidthsMsg{
				Width:      m.width,
				CellWidths: m.cellWidths,
			})
		}

		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:

		m.width = msg.Width
		m.cellWidths = getCellWidths(m.assets)
		for i, r := range m.rows {
			m.rows[i], _ = r.Update(row.SetCellWidthsMsg{
				Width:      m.width,
				CellWidths: m.cellWidths,
			})
		}

		return m, nil

	case row.FrameMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// TODO: send message to a specific row rather than all rows
		for i, r := range m.rows {
			m.rows[i], cmd = r.Update(msg)
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	case ChangeSortMsg:

		var cmd tea.Cmd
		cmds := make([]tea.Cmd, 0)

		// Update the sorter with the new sort option
		m.config.Sort = string(msg)
		m.sorter = s.NewSorter(m.config.Sort)

		// Re-sort and update the assets
		assets := m.sorter(m.assets)
		m.assets = assets

		// Update rows with the new order
		for i, asset := range assets {
			m.rows[i], cmd = m.rows[i].Update(row.UpdateAssetMsg(asset))
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	}

	return m, nil
}

// View rendering hook for bubbletea
func (m *Model) View() string {

	if m.width < 80 {
		return fmt.Sprintf("Terminal window too narrow to render content\nResize to fix (%d/80)", m.width)
	}

	rows := make([]string, 0)
	if m.config.ShowOHLC && len(m.rows) > 0 {
		rows = append(rows, m.renderOHLCHeader())
	}
	for _, row := range m.rows {
		rows = append(rows, row.View())
	}

	return strings.Join(rows, "\n")

}

func (m *Model) renderOHLCHeader() string {
	sortKey := m.config.Sort
	if sortKey == "" {
		sortKey = "change"
	}

	symLabel := "SYMBOL"
	if sortKey == "symbol" || sortKey == "alpha" {
		symLabel += " ▲"
	}

	openLabel := "OPEN"
	if sortKey == "open" {
		openLabel += " ▼"
	}

	highLabel := "HIGH"
	if sortKey == "high" {
		highLabel += " ▼"
	}

	lowLabel := "LOW"
	if sortKey == "low" {
		lowLabel += " ▲"
	}

	closeLabel := "CLOSE"
	if sortKey == "close" {
		closeLabel += " ▼"
	}

	changeLabel := "CHANGE (%)"
	if sortKey == "change" {
		changeLabel += " ▼"
	}

	volumeLabel := "VOLUME"
	if sortKey == "volume" {
		volumeLabel += " ▼"
	}

	mktCapLabel := "MKT CAP"
	if sortKey == "mktcap" || sortKey == "market_cap" {
		mktCapLabel += " ▼"
	}

	headerCells := []grid.Cell{
		{Text: m.config.Styles.TextBold(symLabel), Width: row.WidthName},
		{Text: ""},
		{Text: "", Width: row.WidthMarketState},
		{Text: m.config.Styles.TextBold(openLabel), Width: m.cellWidths.WidthOpen, Align: grid.Right},
		{Text: m.config.Styles.TextBold(highLabel), Width: m.cellWidths.WidthHigh, Align: grid.Right},
		{Text: m.config.Styles.TextBold(lowLabel), Width: m.cellWidths.WidthLow, Align: grid.Right},
		{Text: m.config.Styles.TextBold(closeLabel), Width: m.cellWidths.WidthQuoteExtended, Align: grid.Right},
		{Text: m.config.Styles.TextBold("PRE-MKT"), Width: m.cellWidths.WidthPreMarket, Align: grid.Right},
		{Text: m.config.Styles.TextBold("POST-MKT"), Width: m.cellWidths.WidthPostMarket, Align: grid.Right},
		{Text: m.config.Styles.TextBold(changeLabel), Width: m.cellWidths.WidthQuote, Align: grid.Right},
		{Text: m.config.Styles.TextBold(volumeLabel), Width: m.cellWidths.WidthVolume, Align: grid.Right},
		{Text: m.config.Styles.TextBold(mktCapLabel), Width: m.cellWidths.WidthMarketCap, Align: grid.Right},
	}

	return grid.Render(grid.Grid{
		Rows: []grid.Row{
			{Width: m.width, Cells: headerCells},
			{Width: m.width, Cells: []grid.Cell{{Text: m.config.Styles.TextLine(strings.Repeat("─", m.width))}}},
		},
		GutterHorizontal: row.WidthGutter,
	})
}
func getCellWidths(assets []*c.Asset) row.CellWidthsContainer {

	cellMaxWidths := row.CellWidthsContainer{}

	for _, asset := range assets {
		var quoteLength int

		volumeMarketCapLength := len(u.ConvertFloatToString(asset.QuoteExtended.MarketCap, true))

		if asset.QuoteExtended.FiftyTwoWeekHigh == 0.0 {
			quoteLength = len(u.ConvertFloatToString(asset.QuotePrice.Price, asset.Meta.IsVariablePrecision))
		}

		if asset.QuoteExtended.FiftyTwoWeekHigh != 0.0 {
			quoteLength = len(u.ConvertFloatToString(asset.QuoteExtended.FiftyTwoWeekHigh, asset.Meta.IsVariablePrecision))
		}

		if volumeMarketCapLength > cellMaxWidths.WidthVolumeMarketCap {
			cellMaxWidths.WidthVolumeMarketCap = volumeMarketCapLength
		}

		if quoteLength > cellMaxWidths.QuoteLength {
			cellMaxWidths.QuoteLength = quoteLength
			cellMaxWidths.WidthQuote = quoteLength + row.WidthChangeStatic
			cellMaxWidths.WidthQuoteExtended = quoteLength
			cellMaxWidths.WidthQuoteRange = row.WidthRangeStatic + (quoteLength * 2)
		}

		openLength := len(u.ConvertFloatToString(asset.QuotePrice.PriceOpen, asset.Meta.IsVariablePrecision))
		if openLength > cellMaxWidths.WidthOpen {
			cellMaxWidths.WidthOpen = openLength
		}

		highLength := len(u.ConvertFloatToString(asset.QuotePrice.PriceDayHigh, asset.Meta.IsVariablePrecision))
		if highLength > cellMaxWidths.WidthHigh {
			cellMaxWidths.WidthHigh = highLength
		}

		lowLength := len(u.ConvertFloatToString(asset.QuotePrice.PriceDayLow, asset.Meta.IsVariablePrecision))
		if lowLength > cellMaxWidths.WidthLow {
			cellMaxWidths.WidthLow = lowLength
		}

		volumeLength := len(u.ConvertFloatToString(asset.QuoteExtended.Volume, true))
		if volumeLength > cellMaxWidths.WidthVolume {
			cellMaxWidths.WidthVolume = volumeLength
		}

		if asset.QuoteExtended.MarketCap > 0 {
			mktCapLength := len(u.ConvertFloatToString(asset.QuoteExtended.MarketCap, true))
			if mktCapLength > cellMaxWidths.WidthMarketCap {
				cellMaxWidths.WidthMarketCap = mktCapLength
			}
		}

		if asset.QuotePrice.PricePreMarket > 0 {
			preMktLength := len(u.ConvertFloatToString(asset.QuotePrice.PricePreMarket, asset.Meta.IsVariablePrecision)) + 10
			if preMktLength > cellMaxWidths.WidthPreMarket {
				cellMaxWidths.WidthPreMarket = preMktLength
			}
		}

		if asset.QuotePrice.PricePostMarket > 0 {
			postMktLength := len(u.ConvertFloatToString(asset.QuotePrice.PricePostMarket, asset.Meta.IsVariablePrecision)) + 10
			if postMktLength > cellMaxWidths.WidthPostMarket {
				cellMaxWidths.WidthPostMarket = postMktLength
			}
		}

		if asset.Position != (c.Position{}) {
			positionLength := len(u.ConvertFloatToString(asset.Position.Value, asset.Meta.IsVariablePrecision))
			positionQuantityLength := len(u.ConvertFloatToString(asset.Position.Quantity, asset.Meta.IsVariablePrecision))
			positionUnitCostLength := len(u.ConvertFloatToString(asset.Position.UnitCost, asset.Meta.IsVariablePrecision))

			if positionLength > cellMaxWidths.PositionLength {
				cellMaxWidths.PositionLength = positionLength
				cellMaxWidths.WidthPosition = positionLength + row.WidthChangeStatic + row.WidthPositionGutter
			}

			if positionLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionLength
			}

			if positionQuantityLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionQuantityLength
			}

			if positionUnitCostLength > cellMaxWidths.WidthPositionExtended {
				cellMaxWidths.WidthPositionExtended = positionUnitCostLength
			}

		}

	}

	if cellMaxWidths.WidthOpen < 8 {
		cellMaxWidths.WidthOpen = 8
	}
	if cellMaxWidths.WidthHigh < 8 {
		cellMaxWidths.WidthHigh = 8
	}
	if cellMaxWidths.WidthLow < 8 {
		cellMaxWidths.WidthLow = 8
	}
	if cellMaxWidths.WidthVolume < 8 {
		cellMaxWidths.WidthVolume = 8
	}
	if cellMaxWidths.WidthMarketCap < 10 {
		cellMaxWidths.WidthMarketCap = 10
	}
	if cellMaxWidths.WidthPreMarket < 16 {
		cellMaxWidths.WidthPreMarket = 16
	}
	if cellMaxWidths.WidthPostMarket < 16 {
		cellMaxWidths.WidthPostMarket = 16
	}

	return cellMaxWidths

}
