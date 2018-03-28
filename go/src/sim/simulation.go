package sim

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
)

type ChildConfig struct {
	BirthDate string `json:"birth_date"`
	Tuition []float64 `json:"tuition"`
}

type IRAConfig struct {
	Balance float64 `json:"balance"`
	InheritDate string `json:"inherit_date"`
}

type CarConfig struct {
	PurchaseDate string `json:"purchase_date"`
	Price float64 `json:"price"`
	Principal float64 `json:"principal"`
	Interest float64 `json:"interest"`
	DueDate string `json:"due_date"`
}

type HomeConfig struct {
	Value float64 `json:"value"`
	Principal float64 `json:"principal"`
	Interest float64 `json:"interest"`
	DueDate string `json:"due_date"`
	PropertyTax float64 `json:"property_tax"`
}

type AssetConfig struct {
	Cash float64 `json:"cash"`
	SlushFund map[string]float64 `json:"slush_fund"`
	K401 map[string]float64 `json:"401k"`
	IRA map[string]*IRAConfig `json:"ira"`
	Home *HomeConfig `json:"home"`
	Car *CarConfig `json:"car"`
}

type DebtConfig struct {
	Principal float64 `json:"principal"`
	Interest float64 `json:"interest"`
	DueDate string `json:"due_date"`
}

type RiskProfileConfig struct {
	Speculative float64 `json:"speculative"`
	Aggressive float64 `json:"aggressive"`
	Moderate float64 `json:"moderate"`
	Conservative float64 `json:"conservative"`
}

type AssistedLivingConfig struct {
	BasicRate float64 `json:"basic_rate"`
	TerminalRate float64 `json:"terminal_rate"`
}

type HealthCareConfig struct {
	AssistedLiving *AssistedLivingConfig `json:"assisted_living"`
	Premium float64 `json:"premium"`
	OutOfPocket float64 `json:"out_of_pocket"`
	Inflation float64 `json:"inflation"`
}

type SimConfig struct {
	Spouse *SimConfig `json:"spouse"`
	Name string `json:"name"`
	BirthDate string `json:"birth_date"`
	RiskFactors []string `json:"risk_factors"`
	RetirementAge float64 `json:"retirement_age"`
	SocialSecurityAge float64 `json:"social_security_age"`
	SocialSecurityPayouts [3]float64 `json:"social_security_payouts"`
	HealthCare *HealthCareConfig `json:"health_care"`
	AnnualSalary float64 `json:"annual_salary"`
	MonthlyLiving float64 `json:"monthly_living"`
	Cushion float64 `json:"cushion"`
	Rent float64 `json:"rent"`
	State string `json:"state"`
	Children map[string]*ChildConfig `json:"children"`
	Assets *AssetConfig `json:"assets"`
	RiskProfile *RiskProfileConfig `json:"risk_profile"`
	Debts map[string]*DebtConfig `json:"debts"`
}

type Simulation struct {
	seedPod *SeedPod
	hasRun bool
	startDate time.Time
	config *SimConfig
	Name string
	Spouse *Simulation
	Economy *Economy
	Portfolio *Portfolio
	Actuary *Actuary
	Job *Job
	SocialSecurity *SocialSecurity
	TaxMen *TaxMen
	CashAccount *CashAccount
	HealthCare *HealthCare
	AssistedLiving *AssistedLiving
	Mortgage *Debt
	Home *Home
	Car *Car
	Debts []*Debt
	Investments []*InvestmentAccount
	Children []*Child
	Market float64
	/*
	bc *BusinessCycle
	portfolio *Portfolio
	ssa *SocialSecurity
	car *Car
	house *House
	ira *IRA
	f01k *FourOhOneKay
	slush *SlushFund
	*/
	BalanceHistory []*BalanceData
	Events *EventList
}

func NewSimulation(n int, config *SimConfig) *Simulation {
	seedPod, _ := NewSeedPod("/usr/share/dict/words", int64(n * 1789))
	start := startOfMonth(time.Now())
	s := &Simulation{
		seedPod: seedPod,
		hasRun: false,
		startDate: start,
		config: config,
		Name: config.Name,
	}
	s.Economy = NewEconomy(s)
	s.Portfolio = s.configurePortfolio(config.RiskProfile)
	birthDate, _ := time.ParseInLocation("2006-01-02", config.BirthDate, time.Local)
	s.Actuary = NewActuary(s, birthDate, config.RiskFactors)
	s.CashAccount = NewCashAccount(s, config.Assets.Cash)
	s.TaxMen = NewTaxMen(s, config.State)
	s.Job = NewJob(s, config.AnnualSalary)
	var payout float64 = 0.0
	var age float64 = 62.0
	if config.SocialSecurityAge >= 70.0 {
		payout = config.SocialSecurityPayouts[2]
		age = config.SocialSecurityAge
	} else if config.SocialSecurityAge >= 67.0 {
		payout = config.SocialSecurityPayouts[1]
		age = config.SocialSecurityAge
	} else {
		payout = config.SocialSecurityPayouts[0]
	}
	s.SocialSecurity = NewSocialSecurity(s, age, payout)

	s.HealthCare = s.configureHealthCare(config.HealthCare)
	if config.HealthCare != nil {
		s.AssistedLiving = s.configureAssistedLiving(config.HealthCare.AssistedLiving)
	} else {
		s.AssistedLiving = s.configureAssistedLiving(nil)
	}
	s.Home = s.configureHome(config.Assets.Home, config.Rent)
	s.Car = s.configureCar(config.Assets.Car)
	s.Debts = s.configureDebts(config.Debts)
	s.Investments = s.configureInvestments(config.Assets)
	s.Children = s.configureChildren(config.Children)
	el := EventList([]*Event{})
	s.Events = &el
	if config.Spouse != nil {
		s.Spouse = s.configureSpouse(config.Spouse)
	}
	return s
}

func (s *Simulation) configureSpouse(config *SimConfig) *Simulation {
	sim := &Simulation{
		seedPod: s.seedPod,
		hasRun: false,
		startDate: s.startDate,
		config: config,
		Name: config.Name,
		Economy: s.Economy,
		Portfolio: s.Portfolio,
		CashAccount: s.CashAccount,
		Events: s.Events,
	}
	birthDate, _ := time.ParseInLocation("2006-01-02", config.BirthDate, time.Local)
	var payout float64 = 0.0
	var age float64 = 62.0
	if config.SocialSecurityAge >= 70.0 {
		payout = config.SocialSecurityPayouts[2]
		age = config.SocialSecurityAge
	} else if config.SocialSecurityAge >= 67.0 {
		payout = config.SocialSecurityPayouts[1]
		age = config.SocialSecurityAge
	} else {
		payout = config.SocialSecurityPayouts[0]
	}
	sim.Actuary = NewActuary(sim, birthDate, config.RiskFactors)
	sim.TaxMen = NewTaxMen(sim, s.config.State)
	sim.Job = NewJob(sim, config.AnnualSalary)
	sim.SocialSecurity = NewSocialSecurity(sim, age, payout)
	sim.HealthCare = sim.configureHealthCare(config.HealthCare)
	sim.Car = sim.configureCar(config.Assets.Car)
	return sim
}

func (s *Simulation) configurePortfolio(config *RiskProfileConfig) *Portfolio {
	return NewPortfolio(s, config.Speculative, config.Aggressive, config.Moderate, config.Conservative)
}

func (s *Simulation) configureInvestments(config *AssetConfig) []*InvestmentAccount {
	accounts := []*InvestmentAccount{}
	for name, bal := range config.SlushFund {
		accounts = append(accounts, NewInvestmentAccount(s, bal, name))
	}
	for name, bal := range config.K401 {
		accounts = append(accounts, New401K(s, bal, name))
	}
	for name, cfg := range config.IRA {
		inheritDate, err := time.ParseInLocation("2006-01-02", cfg.InheritDate, time.Local)
		if err != nil {
			inheritDate = s.Actuary.BirthDate
		}
		accounts = append(accounts, NewInheritedIRA(s, cfg.Balance, inheritDate, name))
	}
	return accounts
}

func (s *Simulation) configureDebts(config map[string]*DebtConfig) []*Debt {
	ds := []*Debt{}
	for name, cfg := range config {
		dueDate, err := time.ParseInLocation("2006-01-02", cfg.DueDate, time.Local)
		if err != nil {
			dueDate = time.Now()
		}
		ds = append(ds, NewDebt(s, cfg.Principal, cfg.Interest, dueDate, name))
	}
	return ds
}

func (s *Simulation) configureMortgage(config *HomeConfig) *Debt {
	if config == nil || config.DueDate == "" {
		return NewMortgage(s, 0.0, 0.0, time.Date(2200, time.January, 1, 0, 0, 0, 0, time.Local), "Imaginary")
	}
	dueDate, err := time.ParseInLocation("2006-01-02", config.DueDate, time.Local)
	if err != nil {
		dueDate = time.Now()
	}
	return NewMortgage(s, config.Principal, config.Interest, dueDate, "Home")
}

func (s *Simulation) configureHome(config *HomeConfig, rent float64) *Home {
	mortgage := s.configureMortgage(config)
	if config == nil {
		return NewHome(s, 0.0, mortgage, 0.0, rent)
	}
	return NewHome(s, config.Value, mortgage, config.PropertyTax, rent)
}

func (s *Simulation) configureCarLoan(config *CarConfig) *Debt {
	if config == nil || config.DueDate == "" {
		return NewDebt(s, 0.0, 0.0, time.Date(2200, time.January, 1, 0, 0, 0, 0, time.Local), "Imaginary Car Loan")
	}
	dueDate, err := time.ParseInLocation("2006-01-02", config.DueDate, time.Local)
	if err != nil {
		dueDate = time.Now()
	}
	return NewDebt(s, config.Principal, config.Interest, dueDate, "Car Loan")
}

func (s *Simulation) configureCar(config *CarConfig) *Car {
	loan := s.configureCarLoan(config)
	if config == nil {
		return NewCar(s, 0.0, time.Now(), loan)
	}
	purchaseDate, err := time.ParseInLocation("2006-01-02", config.PurchaseDate, time.Local)
	if err != nil {
		purchaseDate = time.Now()
	}
	return NewCar(s, config.Price, purchaseDate, loan)
}

func (s *Simulation) configureChildren(config map[string]*ChildConfig) []*Child {
	cs := []*Child{}
	for name, cfg := range config {
		bd, err := time.ParseInLocation("2006-01-02", cfg.BirthDate, time.Local)
		if err != nil {
			bd = time.Now()
		}
		cs = append(cs, NewChild(s, name, bd, cfg.Tuition))
	}
	return cs
}

func (s *Simulation) configureAssistedLiving(config *AssistedLivingConfig) *AssistedLiving {
	if config == nil {
		return NewAssistedLiving(s, 0.0, 0.0)
	}
	return NewAssistedLiving(s, config.BasicRate, config.TerminalRate)
}

func (s *Simulation) configureHealthCare(config *HealthCareConfig) *HealthCare {
	if config == nil {
		return NewHealthCare(s, 0.0, 0.0, 0.0)
	}
	return NewHealthCare(s, config.Premium, config.OutOfPocket, config.Inflation)
}

func (s *Simulation) Rng() (*rand.Rand, error) {
	return s.seedPod.Next()
}

func (s *Simulation) StartDate() time.Time {
	return s.startDate
}

func (s *Simulation) RetirementAge() float64 {
	return s.config.RetirementAge
}

func (s *Simulation) RetirementDate() time.Time {
	yf := s.RetirementAge()
	yi := int(yf)
	d := int(365.25 * (yf - float64(yi)))
	return s.Actuary.BirthDate.AddDate(yi, 0, d)
}

func (s *Simulation) SocialSecurityAge() float64 {
	return s.config.SocialSecurityAge
}

func (s *Simulation) run() {
	date := s.StartDate()
	var months float64 = 0.0
	s.Market = 1.0
	s.BalanceHistory = []*BalanceData{}
	busted := false
	var surplus float64 = 0.0
	for date.Before(s.Actuary.DeathDate) {
		prev := s.CashAccount.Balance()
		haveHome := s.Home != nil && s.Home.Value() > 0.0
		s.CashAccount.AccrueMarketReturn(date)
		if date.Month() == time.January {
			s.TaxMen.Annual(date)
		}
		s.Job.Monthly(date)
		s.SocialSecurity.Monthly(date)
		if s.Home != nil {
			s.Home.Monthly(date)
		}
		if s.Car != nil {
			s.Car.Monthly(date)
		}
		for _, c := range s.Children {
			c.Monthly(date)
		}
		for _, debt := range s.Debts {
			debt.Monthly(date)
		}
		s.HealthCare.Monthly(date)
		s.AssistedLiving.Monthly(date)
		s.CashAccount.Withdraw(s.config.MonthlyLiving, date, "Monthly Expenses")
		if s.Spouse != nil && date.Before(s.Spouse.Actuary.DeathDate) {
			if date.Month() == time.January {
				s.Spouse.TaxMen.Annual(date)
			}
			s.Spouse.Job.Monthly(date)
			s.Spouse.SocialSecurity.Monthly(date)
			s.Spouse.HealthCare.Monthly(date)
			s.Car.Monthly(date)
		}
		if date.Month() == time.December {
			for _, acct := range s.Investments {
				acct.TakeRMD(date)
			}
		}
		surplus += (s.CashAccount.Balance() - prev)

		if surplus > 0.0 && date.Month() == s.StartDate().Month() && !date.Equal(s.StartDate()) {
			// move cash to investments
			for _, acct := range s.Investments {
				if !acct.Taxable() && acct.CanDeposit(date) {
					acct.Deposit(surplus, date, "Invest")
					surplus = 0.0
					break
				}
			}
		}

		// top off cash account from invetments, favoring those with
		// the largest RMDs
		for s.CashAccount.Balance() < s.config.Cushion  * 0.75 {
			tgt := s.config.Cushion - s.CashAccount.Balance()
			var maxRmd float64
			var maxRmdAcct *InvestmentAccount
			for _, acct := range s.Investments {
				rmd := acct.RMD(date)
				if rmd > 0.0 {
					if maxRmdAcct == nil || rmd > maxRmd {
						maxRmd = rmd
						maxRmdAcct = acct
					}
				}
			}
			if maxRmdAcct == nil {
				break
			}
			if tgt > maxRmd {
				tgt = maxRmd
			}
			maxRmdAcct.Withdraw(tgt, date, "Cushion")
		}

		// top off cash account from investments, favoring untaxed
		for s.CashAccount.Balance() < s.config.Cushion * 0.75 {
			tgt := s.config.Cushion - s.CashAccount.Balance()
			found := false
			for _, acct := range s.Investments {
				if !acct.CanWithdraw(date) {
					continue
				}
				if acct.Taxable() {
					continue
				}
				if acct.Balance() > 0.0 {
					found = true
					acct.Withdraw(tgt, date, "Cushion")
					break
				}
			}
			if !found {
				break
			}
		}
		for s.CashAccount.Balance() < s.config.Cushion * 0.75 {
			tgt := s.config.Cushion - s.CashAccount.Balance()
			found := false
			for _, acct := range s.Investments {
				if !acct.CanWithdraw(date) {
					continue
				}
				if acct.Balance() > 0.0 {
					found = true
					acct.Withdraw(tgt, date, "Cushion")
					break
				}
			}
			if !found {
				break
			}
		}

		// top off cash account by selling house
		if s.CashAccount.Balance() < s.config.Cushion * 0.25 {
			if s.Home.Balance() > 0.0 {
				s.Events.Add(date, "Liquidity Crisis", 1)
				s.Home.Liquidate(date)
			}
		}

		if haveHome && !(s.Home != nil && s.Home.Value() > 0.0) {
			minBalance := math.Max(prev, s.config.Cushion * 1.5)
			amt := s.CashAccount.Balance() - minBalance
			if amt > 0.0 {
				for _, acct := range s.Investments {
					if !acct.Taxable() && acct.CanDeposit(date) {
						acct.Deposit(amt, date, "ReInvest")
						amt = 0.0
						break
					}
				}
			}
		}

		for _, acct := range s.Investments {
			acct.AccrueMarketReturn(date)
		}
		s.Market *= (1.0 + s.Economy.MarketReturn(date) / 1200.0)
		s.CashAccount.Reconcile()
		s.BalanceHistory = append(s.BalanceHistory, s.Balances(date))
		if !busted && s.Balance() <= 0.0 {
			s.Events.Add(date, "Bankruptcy", 10)
			busted = true
		}
		date = incrementMonth(date)
		months += 1.0
	}
	s.Market = math.Pow(s.Market, 12.0 / months) - 1.0
	if s.RetirementDate().Before(s.Actuary.DeathDate) {
		s.Events.Add(s.RetirementDate(), "Retire", 6)
	}
	if s.SocialSecurity.BenefitsDate().Before(s.Actuary.DeathDate) {
		s.Events.Add(s.SocialSecurity.BenefitsDate(), "Social Security", 6)
	}
	for _, child := range s.Children {
		if child.GraduationDate().Before(s.Actuary.DeathDate) {
			s.Events.Add(child.GraduationDate(), child.Name() + " Graduation", 6)
		}
	}
	s.Events.Add(s.Actuary.DeathDate, "Death", 10)
	if s.Spouse != nil && s.Spouse.Actuary.DeathDate.Before(s.Actuary.DeathDate) {
		s.Events.Add(s.Spouse.Actuary.DeathDate, s.Spouse.Name + "'s Death", 10)
	}
	sort.Sort(s.Events)
}

func (s *Simulation) Balance() float64 {
	var balance float64 = 0.0
	if s.Home != nil {
		balance += s.Home.Balance()
	}
	if s.Car != nil {
		car := s.Car.Balance()
		if car > 0.0 {
			car /= 2.0
		}
		balance += car
	}
	for _, debt := range s.Debts {
		balance += debt.Balance()
	}
	for _, acct := range s.Investments {
		balance += acct.Balance()
	}
	balance += s.CashAccount.Balance()
	return balance
}

type BalanceData struct {
	Date string `json:"date"`
	Balances map[string]float64 `json:"balances"`
}

func (s *Simulation) Balances(date time.Time) *BalanceData {
	date = endOfMonth(date)
	bd := &BalanceData{
		Date: date.Format("2006-01-02"),
		Balances: map[string]float64{},
	}
	if s.Home != nil {
		bal := round2(s.Home.Balance())
		if bal != 0.0 {
			bd.Balances["Home"] = bal
		}
		if s.Home.Mortgage() != nil {
			bal := round2(s.Home.Mortgage().Balance())
			if bal != 0.0 {
				bd.Balances[s.Home.Mortgage().Name()] = bal
			}
		}
	}
	if s.Car != nil {
		bal := round2(s.Car.Balance())
		if bal != 0.0 {
			bd.Balances[s.Name + "'s Car"] = bal
		}
		if s.Car.Loan() != nil {
			bal := round2(s.Car.Loan().Balance())
			if bal != 0.0 {
				bd.Balances[s.Name + "'s Car Loan"] = bal
			}
		}
	}
	for i, debt := range s.Debts {
		bal := round2(debt.Balance())
		if bal != 0.0 {
			name := fmt.Sprintf("Debt %d - %s", i, debt.Name())
			bd.Balances[name] = bal
		}
	}
	for i, acct := range s.Investments {
		bal := round2(acct.Balance())
		if bal != 0.0 {
			name := fmt.Sprintf("Acct %d - %s", i, acct.Name())
			bd.Balances[name] = bal
		}
	}
	bd.Balances["Cash"] = round2(s.CashAccount.Balance())
	bd.Balances["Total"] = round2(s.Balance())
	return bd
}

func (s *Simulation) Liabilities(date time.Time) float64 {
	var l float64 = 0.0
	for _, c := range s.Children {
		l += c.Remaining(date)
	}
	return l
}

type Results struct {
	Index int `json:"index"`
	Age float64 `json:"death"`
	Balance float64 `json:"balance"`
	Market float64 `json:"market"`
	Transactions *Ledger `json:"transactions,omitempty"`
	Balances []*BalanceData `json:"balances,omitempty"`
	Events *EventList `json:"events,omitempty"`
}

func (s *Simulation) Results() *Results {
	if !s.hasRun {
		s.run()
		s.hasRun = true
	}
	balance := s.Balance()
	lia := s.Liabilities(s.Actuary.DeathDate)
	return &Results{
		Age: s.Actuary.DeathAge(),
		Balance: balance - lia,
		Market: s.Market,
		Transactions: s.CashAccount.Transactions(nil, nil),
		Balances: s.BalanceHistory,
		Events: s.Events,
	}
}
