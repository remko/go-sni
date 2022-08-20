package sni

import (
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

type Item struct {
	name  string
	conn  *dbus.Conn
	props *prop.Properties

	// Caled when activation is requested for the item (e.g. clicked)
	Activate func(x, y int)

	// Called when secondary activation is requested for the item (e.g. middle-click)
	SecondaryActivate func(x, y int)

	// Called when a context menu is requested for the item (e.g. right-click)
	ContextMenu func(x, y int)

	// Called when scroll is requested for the item (e.g. scroll wheel)
	Scroll func(delta int, direction string)
}

type Icon struct {
	Name    string
	Pixmaps []Pixmap
}

// See https://www.freedesktop.org/wiki/Specifications/StatusNotifierItem/StatusNotifierItem/ for details of tese fields
type ItemConfig struct {
	Category      string
	ID            string
	Title         string
	Status        string
	Icon          Icon
	OverlayIcon   Icon
	AttentionIcon Icon
	Tooltip       Tooltip
}

type Tooltip struct {
	Icon  Icon
	Title string
	Text  string
}

type notifierItem struct{ it *Item }

func (ni *notifierItem) ContextMenu(x, y int) *dbus.Error {
	if ni.it.ContextMenu != nil {
		ni.it.ContextMenu(x, y)
	}
	return nil
}

func (ni *notifierItem) Activate(x, y int) *dbus.Error {
	if ni.it.Activate != nil {
		ni.it.Activate(x, y)
	}
	return nil
}

func (ni *notifierItem) SecondaryActivate(x, y int) *dbus.Error {
	if ni.it.SecondaryActivate != nil {
		ni.it.SecondaryActivate(x, y)
	}
	return nil
}

func (ni *notifierItem) Scroll(delta int, direction string) *dbus.Error {
	if ni.it.Scroll != nil {
		ni.it.Scroll(delta, direction)
	}
	return nil
}

func NewItem(conf ItemConfig) (*Item, error) {
	var err error
	sni := &Item{}
	sni.name = fmt.Sprintf("org.freedesktop.StatusNotifierItem-%d-1", os.Getpid())
	sni.conn, err = dbus.ConnectSessionBus()
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			sni.conn.Close()
		}
	}()

	// Register service
	resp, err := sni.conn.RequestName(sni.name, dbus.NameFlagDoNotQueue)
	if err != nil {
		return nil, err
	}
	if resp != dbus.RequestNameReplyPrimaryOwner {
		return nil, fmt.Errorf("name %s already taken", sni.name)
	}

	// Export
	ni := notifierItem{it: sni}
	if err := sni.conn.Export(&ni, "/StatusNotifierItem", "org.freedesktop.StatusNotifierItem"); err != nil {
		return nil, err
	}

	propss := map[string]map[string]*prop.Prop{
		"org.freedesktop.StatusNotifierItem": {
			"Category":            {Value: conf.Category},
			"Id":                  {Value: conf.ID},
			"Title":               {Value: conf.Title},
			"Status":              {Value: conf.Status},
			"WindowId":            {Value: 0},
			"IconName":            {Value: conf.Icon.Name},
			"IconPixmap":          {Value: conf.Icon.Pixmaps},
			"OverlayIconName":     {Value: conf.OverlayIcon.Name},
			"OverlayIconPixmap":   {Value: conf.OverlayIcon.Pixmaps},
			"AttentionIconName":   {Value: conf.AttentionIcon.Name},
			"AttentionIconPixmap": {Value: conf.AttentionIcon.Pixmaps},
			"AttentionMovieName":  {Value: ""},
			"ToolTip":             {Value: conf.Tooltip},
			"ItemIsMenu":          {Value: false},
			"Menu":                {Value: dbus.ObjectPath("/MenuBar")},
		},
	}
	sni.props, err = prop.Export(sni.conn, "/StatusNotifierItem", propss)
	if err != nil {
		return nil, err
	}

	// Introspection
	sni.conn.Export(introspect.NewIntrospectable(&introspect.Node{Children: []introspect.Node{{Name: "StatusNotifierItem"}}}), "/", "org.freedesktop.DBus.Introspectable")
	sni.conn.Export(introspect.NewIntrospectable(&introspect.Node{
		Interfaces: []introspect.Interface{
			{
				Name:       "org.freedesktop.StatusNotifierItem",
				Methods:    introspect.Methods(&ni),
				Properties: sni.props.Introspection("org.freedesktop.StatusNotifierItem"),
				Signals: []introspect.Signal{
					{Name: "NewTitle"},
					{Name: "NewIcon"},
					{Name: "NewAttentionIcon"},
					{Name: "NewOverlayIcon"},
					{Name: "NewToolTip"},
					{Name: "NewStatus", Args: []introspect.Arg{{Name: "status", Type: "s"}}},
				},
			},
		},
	}), "/StatusNotifierItem", "org.freedesktop.DBus.Introspectable")

	// Register with watcher
	obj := sni.conn.Object("org.freedesktop.StatusNotifierWatcher", "/StatusNotifierWatcher").Call("org.freedesktop.StatusNotifierWatcher.RegisterStatusNotifierItem", dbus.FlagNoReplyExpected, sni.name)
	if obj.Err != nil {
		return nil, obj.Err
	}

	success = true
	return sni, nil
}

func (sni *Item) SetTitle(t string) error {
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "Title", t)
	sni.conn.Emit("/StatusNotifierItem", "org.freedesktop.StatusNotifierItem.NewTitle")
	return nil
}

func (sni *Item) SetIcon(icon Icon) error {
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "IconName", icon.Name)
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "Pixmap", icon.Pixmaps)
	sni.conn.Emit("/StatusNotifierItem", "org.freedesktop.StatusNotifierItem.NewIcon")
	return nil
}

func (sni *Item) SetOverlayIcon(icon Icon) error {
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "OverlayIconName", icon.Name)
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "OverlayPixmap", icon.Pixmaps)
	sni.conn.Emit("/StatusNotifierItem", "org.freedesktop.StatusNotifierItem.NewOverlayIcon")
	return nil
}

func (sni *Item) SetAttentionIcon(icon Icon) error {
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "AttentionIconName", icon.Name)
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "AttentionPixmap", icon.Pixmaps)
	sni.conn.Emit("/StatusNotifierItem", "org.freedesktop.StatusNotifierItem.NewAttentionIcon")
	return nil
}

func (sni *Item) SetTooltip(tooltip Tooltip) error {
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "ToolTip", tooltip)
	sni.conn.Emit("/StatusNotifierItem", "org.freedesktop.StatusNotifierItem.NewToolTip")
	return nil
}

func (sni *Item) SetStatus(status string) error {
	sni.props.SetMust("org.freedesktop.StatusNotifierItem", "Status", status)
	sni.conn.Emit("/StatusNotifierItem", "org.freedesktop.StatusNotifierItem.NewStatus", status)
	return nil
}

func (sni *Item) Close() error {
	return sni.conn.Close()
}
