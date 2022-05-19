package trend

import (
	"github.com/thecolngroup/zerotoalgo/internal/dec"
	"github.com/thecolngroup/zerotoalgo/internal/util"
	"github.com/thecolngroup/zerotoalgo/market"
	"github.com/thecolngroup/zerotoalgo/money"
	"github.com/thecolngroup/zerotoalgo/risk"
	"github.com/thecolngroup/zerotoalgo/ta"
	"github.com/thecolngroup/zerotoalgo/trader"
)

var _ trader.MakeFromConfig = MakeApexBotFromConfig

// MakeApexBotFromConfig returns a bot configured with an ApexPredicter.
func MakeApexBotFromConfig(config map[string]any) (trader.Bot, error) {

	var bot Bot

	bot.Asset = market.NewAsset(util.ToString(config["asset"]))

	bot.EnterLong = util.ToFloat(config["enterlong"])
	bot.EnterShort = util.ToFloat(config["entershort"])
	bot.ExitLong = util.ToFloat(config["exitlong"])
	bot.ExitShort = util.ToFloat(config["exitshort"])

	maLength := util.ToInt(config["malength"])
	ma := ta.NewALMA(maLength)
	mmi := ta.NewMMI(util.ToInt(config["mmilength"]))
	bot.Predicter = NewApexPredicter(ma, mmi)

	riskSDLength := util.ToInt(config["riskersdlength"])
	if riskSDLength > 0 {
		bot.Risker = risk.NewSDRisker(riskSDLength, util.ToFloat(config["riskersdfactor"]))
	} else {
		bot.Risker = risk.NewFullRisker()
	}

	initialCapital := dec.New(util.ToFloat(config["initialcapital"]))
	sizerF := util.ToFloat(config["sizerf"])
	if sizerF > 0 {
		bot.Sizer = money.NewSafeFSizer(initialCapital, sizerF, util.ToFloat(config["sizerscalef"]))
	} else {
		bot.Sizer = money.NewFixedSizer(initialCapital)
	}

	return &bot, nil
}
