package signaldbus

// signalbot
// Copyright (C) 2024  Lukas Heindl
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

///////////////////////////
// GROUP RELATED METHODS //
///////////////////////////

// Create a new group.
// groupName is the name of the new group, members is a string array of the new
// members to invite, avatar is the filename of an avatar-picture for this
// group (empty if none). The returned groupId is the identification of the new
// group.
// Might raise `AttachmentInvalid`, `Failure`, `InvalidNumber` exceptions.
func (s *Account) CreateGroup(groupName string, members []string, avatar string) (groupId []byte, err error) {
	// TODO arg members: phone numbers or names?
	call := s.obj.Call("org.asamk.Signal.createGroup", 0, groupName, members, avatar)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.AttachmentInvalid":
			return []byte{}, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return []byte{}, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidNumber":
			return []byte{}, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&groupId); err != nil {
		return []byte{}, fmt.Errorf("signal-cli: %v", err)
	}
	return groupId, nil
}

// TODO how to return an invalid objectpath and we shouldn't expose dbus stuff here
// func (s *Account) GetGroup(groupId []byte) (objectPath dbus.ObjectPath, err error) {
// 	call := s.obj.Call("org.asamk.Signal.getGroup", 0, groupId)
// 	if call.Err != nil {
// 		err := call.Err.(dbus.Error) // panics if assertion does not succeed
// 		switch err.Name {
// 			default:
// 				panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
// 		}
// 	}
// 	if err := call.Store(&objectPath); err != nil {
// 		return nil, fmt.Errorf("signal-cli: %v", err)
// 	}
// 	return objectPath, nil
// }

// Translate a groupId to the name of the group.
// Might raise InvalidGroupId exception
func (s *Account) GetGroupName(groupId []byte) (name string, err error) {
	call := s.obj.Call("org.asamk.Signal.getGroupName", 0, groupId)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidGroupId":
			return "", fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&name); err != nil {
		return "", fmt.Errorf("signal-cli: %v", err)
	}
	return name, nil
}

// Get the members of a group.
// groupId identifies the group to query. The returned members is a string
// array with the phoneNumbers of all active members (if the group wasn't found
// this array is empty).
func (s *Account) GetGroupMembers(groupId []byte) (members []string, err error) {
	call := s.obj.Call("org.asamk.Signal.getGroupMembers", 0, groupId)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&members); err != nil {
		return []string{}, fmt.Errorf("signal-cli: %v", err)
	}
	return members, nil
}

// Join a group with an inviteURI.
// inviteURI is the URI of the invitation.
// Might raise `Failure` exception.
func (s *Account) JoinGroup(inviteURI string) (err error) {
	call := s.obj.Call("org.asamk.Signal.joinGroup", 0, inviteURI)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// TODO how to return an invalid objectpath and we shouldn't expose dbus stuff here
// func (s *Account) ListGroups() (groups<a(oays)>, err error) {
// 	listGroups
// }

// Send a message to a group.
// message is the test to send (can be UTF8), attachments is a string array of
// filenames to attach (files must be accessible for signal-cli), groupID
// identifies the group to send to. The returned timestamp can be used to
// identify the corresponding signal reply.
// Might raise `GroupNotFound`, `Failure`, `AttachmentInvalid`, `InvalidGroupId` exceptions.`
func (s *Account) SendGroupMessage(message string, attachments []string, groupId []byte) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendGroupMessage", 0, message, attachments, groupId)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.GroupNotFound":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.AttachmentInvalid":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidGroupId":
			return 0, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&timestamp); err != nil {
		return 0, fmt.Errorf("signal-cli: %v", err)
	}
	return timestamp, nil
}

// Set typing indicator for a group.
// groupId identifies the group to send to, if stop is true the typing
// indicator is stopped.
// Might raise `Failure`, `GroupNotFound`, `UntrustedIdentity` exceptions.
func (s *Account) SendGroupTyping(groupId []byte, stop bool) (err error) {
	call := s.obj.Call("org.asamk.Signal.sendGroupTyping", 0, groupId, stop)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.GroupNotFound":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.UntrustedIdentity":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Send a reaction to a message in a group.
// emoji is the emoji to react with, if remove is true a previous reaction is
// removed, targetAuthor is the phoneNumber of the author of the message to
// react to, targetSentTimestamp identifies the message to work on, groupID
// identifies the group wor work on.
// Might raise `Failure`, `InvalidNumber`, `GroupNotFound`, `InvalidGroupId` exceptions.
func (s *Account) SendGroupMessageReaction(emoji string, remove bool, targetAuthor string, targetSentTimestamp int64, groupId []byte) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendGroupMessageReaction", 0, emoji, remove, targetAuthor, targetSentTimestamp, groupId)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.Failure":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidNumber":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.GroupNotFound":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidGroupId":
			return 0, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&timestamp); err != nil {
		return 0, fmt.Errorf("signal-cli: %v", err)
	}
	return timestamp, nil
}

// Delete a message in a group.
// targetSentTimestamp identifies the message to delete, groupId identifies the
// group to work on. The returned Timestamp can be used to identify the correspnding signal reply.
// Might raise `Failure`, `GroupNotFound`, `InvalidGroupId` exceptions.
func (s *Account) SendGroupRemoteDeleteMessage(targetSentTimestamp int64, groupId []byte) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendGroupRemoteDeleteMessage", 0, targetSentTimestamp, groupId)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.Failure":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.GroupNotFound":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidGroupId":
			return 0, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&timestamp); err != nil {
		return 0, fmt.Errorf("signal-cli: %v", err)
	}
	return timestamp, nil
}
