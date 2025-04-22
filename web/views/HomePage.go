package views

import (
	"net/http"
)

type BasePage struct {
	Title        string
	BodyClass    string
	CustomHeader bool
}

type HomePage struct {
	BasePage
	Header         Header
	DesignListView DesignListView
	ShowSearch     bool
}

func (p *HomePage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.Header.Load(r, w, vc)
	p.DesignListView.Load(r, w, vc)
	// p.ShowSearch = true
	return
}

type BrowsePage struct {
	Header Header
}

func (p *BrowsePage) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	p.Header.Load(r, w, vc)
	return
}

type PrivacyPolicy struct {
	Header Header
}

func (p *PrivacyPolicy) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	return p.Header.Load(r, w, vc)
}

type TermsOfService struct {
	Header Header
}

func (p *TermsOfService) Load(r *http.Request, w http.ResponseWriter, vc *ViewContext) (err error, finished bool) {
	return p.Header.Load(r, w, vc)
}

func (g *TermsOfService) Copy() View { return &TermsOfService{} }
func (g *PrivacyPolicy) Copy() View  { return &PrivacyPolicy{} }
func (g *HomePage) Copy() View       { return &HomePage{} }
func (g *BrowsePage) Copy() View     { return &BrowsePage{} }
