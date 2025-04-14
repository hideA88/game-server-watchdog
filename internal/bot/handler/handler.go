package handler

import "github.com/bwmarrin/discordgo"

type Handler interface {
	Handle(s *discordgo.Session, m *discordgo.MessageCreate)
}
