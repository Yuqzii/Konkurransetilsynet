package codeforces

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// @abstract	Creates a channel in every guild if it does not already have one
// @return		Slice of channel IDs, one for each guild. Includes preexisting channels with the name.
func createChannelIfNotExist(s *discordgo.Session, channelName string, guilds []*discordgo.Guild) (result []string, err error) {
	for _, guild := range guilds {
		channel, err := getChannelIDByName(channelName, guild.ID, s)
		if err != nil {
			return nil, fmt.Errorf("getting channel ID of '%s' in guild %s: %w", channelName, guild.ID, err)
		}

		// Create new channel if there does not exist one
		if channel == "" {
			newChannel, err := s.GuildChannelCreate(guild.ID, channelName, discordgo.ChannelTypeGuildText)
			if err != nil {
				return nil, fmt.Errorf("creating channel '%s' in guild %s: %w", channelName, guild.ID, err)
			}
			channel = newChannel.ID
		}

		result = append(result, channel)
	}

	return result, nil
}

// @return	ID of the channel as a string, empty ("") if there is no channel with the provided name.
func getChannelIDByName(name string, guildID string, s *discordgo.Session) (string, error) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		return "", err
	}

	// Try to find a channel with the name
	for _, channel := range channels {
		// Skip non-text channels
		if channel.Type != discordgo.ChannelTypeGuildText {
			continue
		}

		if channel.Name == name {
			return channel.ID, nil
		}
	}

	return "", nil
}

// @abstract	Creates a role in every guild if it does not already have one.
// @return		Slice of role IDs, one for each guild. Includes preexisting roles with the name.
func createRoleIfNotExists(s *discordgo.Session, roleName string, guilds []*discordgo.Guild) (result []string, err error) {
	for _, guild := range guilds {
		role, err := getRoleIDByName(roleName, guild.ID, s)
		if err != nil {
			return nil, fmt.Errorf("getting role ID of '%s' in guild %s: %w", roleName, guild.ID, err)
		}

		// Create new role if there does not exist one
		if role == "" {
			newRole, err := s.GuildRoleCreate(guild.ID, &discordgo.RoleParams{
				Name: roleName,
			})
			if err != nil {
				return nil, fmt.Errorf("creating role '%s' in guild %s: %w", roleName, guild.ID, err)
			}

			role = newRole.ID
		}

		result = append(result, role)
	}

	return result, nil
}

// @return	ID of the role as a string, empty ("") if there is no role with the provided name.
func getRoleIDByName(name string, guildID string, s *discordgo.Session) (string, error) {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		return "", err
	}

	// Try to find role with correct name
	for _, role := range roles {
		if role.Name == name {
			return role.ID, nil
		}
	}

	return "", nil
}
