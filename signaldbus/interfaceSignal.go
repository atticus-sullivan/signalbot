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

//////////////////////
// SIGNAL INTERFACE //
//////////////////////

// Get the name correspnding to a number.
func (s *Account) GetContactName(number string) (name string, err error) {
	call := s.obj.Call("org.asamk.Signal.getContactName", 0, number)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&name); err != nil {
		return "", fmt.Errorf("signal-cli: %v", err)
	}
	return name, nil
}

// Searches contacts and known profiles for a given name and returns the list
// of all known numbers. May result in e.g. two entries if a contact and
// profile name is set.
func (s *Account) GetContactNumber(name string) (numbers []string, err error) {
	call := s.obj.Call("org.asamk.Signal.getContactNumber", 0, name)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&numbers); err != nil {
		return []string{}, fmt.Errorf("signal-cli: %v", err)
	}
	return numbers, nil
}

// For unknown numbers false is returned but no exception is raised.
func (s *Account) GetSelfNumber() (number string, err error) {
	call := s.obj.Call("org.asamk.Signal.getSelfNumber", 0)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&number); err != nil {
		return "", fmt.Errorf("signal-cli: %v", err)
	}
	return number, nil
}

// For unknown numbers false is returned but no exception is raised.
// Might rise `InvalidPhoneNumber` exception
func (s *Account) IsContactBlocked(number string) (blocked bool, err error) {
	call := s.obj.Call("org.asamk.Signal.isContactBlocked", 0, number)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidPhoneNumber":
			return false, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&blocked); err != nil {
		return false, fmt.Errorf("signal-cli: %v", err)
	}
	return blocked, nil
}

// Check if we are registered, false is returned, but no exception is raised.
// If no number is given, returns true (indicating that you are registered).
// Might rise `InvalidPhoneNumber` exception
func (s *Account) IsRegistered() (result bool, err error) {
	call := s.obj.Call("org.asamk.Signal.isRegistered", 0)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidPhoneNumber":
			return false, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&result); err != nil {
		return false, fmt.Errorf("signal-cli: %v", err)
	}
	return result, nil
}

// For unknown number, false is returned, but no exception is raised. If no
// number is given, returns true (indicating that you are registered).
// Might rise `InvalidPhoneNumber` exception
func (s *Account) IsRegistered_num(number string) (result bool, err error) {
	call := s.obj.Call("org.asamk.Signal.isRegistered", 0, number)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidPhoneNumber":
			return false, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&result); err != nil {
		return false, fmt.Errorf("signal-cli: %v", err)
	}
	return result, nil
}

// For unknown numbers, false is returned, but no exception is raised. If no
// number is given, returns true (indicating that you are registered).
// Might rise `InvalidPhoneNumber` exception
func (s *Account) IsRegistered_nums(numbers []string) (results []bool, err error) {
	call := s.obj.Call("org.asamk.Signal.isRegistered", 0, numbers)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidPhoneNumber":
			return []bool{}, fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&results); err != nil {
		return []bool{}, fmt.Errorf("signal-cli: %v", err)
	}
	return results, nil
}

// This is a concatenated list of all defined contacts as well of profiles
// known (e.g. peer group members or sender of received messages)
func (s *Account) ListNumbers() (numbers []string, err error) {
	call := s.obj.Call("org.asamk.Signal.listNumbers", 0)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&numbers); err != nil {
		return []string{}, fmt.Errorf("signal-cli: %v", err)
	}
	return numbers, nil
}

// Removes registration PIN protection.
// Might raise `Failure` exception
func (s *Account) RemovePin() (err error) {
	call := s.obj.Call("org.asamk.Signal.removePin", 0)
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

// Might raise `Failure`, `InvalidNumber`, `UntrustedIdentity` exceptions
func (s *Account) SendEndSessionMessage(recipients []string) (err error) {
	call := s.obj.Call("org.asamk.Signal.sendEndSessionMessage", 0, recipients)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidNumber":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.UntrustedIdentity":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Sends a message to one recipient.
// message is the text to send (can be UTF-8), attachments is a string array of
// filenames to send as attachments (process needs to have access to these
// files), recipient is the number of the single-recipient. The returned
// timestamp can be used to identify the corresponding signal reply.
// Might raise `AttachmentInvalid`, `Failure`, `InvalidNumber`, `UntrustedIdentity` exceptions.`
func (s *Account) SendMessage(message string, attachments []string, recipient string, notifySelf bool) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendMessage", 0, message, attachments, recipient, notifySelf)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.AttachmentInvalid":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidNumber":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.UntrustedIdentity":
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

// Sends a message to multiple recipients.
// message is the text to send (can be UTF-8), attachments is a string array of
// filenames to send as attachments (process needs to have access to these
// files), recipients is the string array of the numbers of the recipients. The
// returned timestamp can be used to identify the corresponding signal reply.
// Might raise `AttachmentInvalid`, `Failure`, `InvalidNumber`, `UntrustedIdentity` exceptions.`
func (s *Account) SendMessage_multi(message string, attachments []string, recipients []string, notifySelf bool) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendMessage", 0, message, attachments, recipients, notifySelf)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.AttachmentInvalid":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidNumber":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.UntrustedIdentity":
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

// Sends a reaction to a message.
// emoji is the unicode grapheme cluster of the emoji, remove is a boolean
// whether a previously set emoji should be removed, targetAuthor is the phone
// number of the author of the messate to which react to, targetSentTimestamp
// is the timestamp of the message to react to, recipient is the phoneNumber of
// the recipient to send the reaction to. The returned timestamp can be used to
// identify the correspnding signal reply.
// Might raise `Failure`, `InvalidNumber` exceptions.`
func (s *Account) SendMessageReaction(emoji string, remove bool, targetAuthor string, targetSentTimestamp int64, recipient string) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendMessageReaction", 0, emoji, remove, targetAuthor, targetSentTimestamp, recipient)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.Failure":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.InvalidNumber":
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

// Sends a reaction to a message.
// emoji is the unicode grapheme cluster of the emoji, remove is a boolean
// whether a previously set emoji should be removed, targetAuthor is the phone
// number of the author of the message to which react to, targetSentTimestamp
// is the timestamp of the message to react to, recipients is the string array
// of phoneNumbers of the recipients to send the reaction to. The returned
// timestamp can be used to identify the correspnding signal reply.
// Might raise `Failure`, `InvalidNumber` exceptions.`
func (s *Account) SendMessageReaction_multi(emoji string, remove bool, targetAuthor string, targetSentTimestamp int64, recipients []string) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendMessageReaction", 0, emoji, remove, targetAuthor, targetSentTimestamp, recipients)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&timestamp); err != nil {
		return 0, fmt.Errorf("signal-cli: %v", err)
	}
	return timestamp, nil
}

// func (s *Account) SendPaymentNotification(receipt []byte, note string, recipient string) (timestamp int64, err error) {
// 	sendPaymentNotification
// }

// Sends a message to the noteToSelf chat.
// message is the text to send (can be UTF-8), attachments is a string array of
// filenames to send as attachments (process needs to have access to these
// files), The returned timestamp can be used to identify the corresponding
// signal reply.
// Might raise `AttachmentInvalid`, `Failure` exceptions.`
func (s *Account) SendNoteToSelfMessage(message string, attachments []string) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendNoteToSelfMessage", 0, message, attachments)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.AttachmentInvalid":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
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

// Sends a READ receipt. Usually this is sent when the users sees message. See
// https://github.com/AsamK/signal-cli/discussions/948
// recipient is the phoneNumber to send the read-receipt to,
// targetSentTimestamps is the array to identify the corresponding
// signal-messages.
// Might raise `Failure`, `UntrustedIdentity` exceptions.
func (s *Account) SendReadReceipt(recipient string, targetSentTimestamps []int64) (err error) {
	call := s.obj.Call("org.asamk.Signal.sendReadReceipt", 0, recipient, targetSentTimestamps)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.UntrustedIdentity":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Sends a VIEWED receipt. Usually this is sent when the users e.g. listens to
// a voice message. See https://github.com/AsamK/signal-cli/discussions/948
// recipient is the phoneNumber to send the viewed-receipt to,
// targetSentTimestamps is the array to identify the corresponding
// signal-messages.
// Might raise `Failure`, `UntrustedIdentity` exceptions.
func (s *Account) SendViewedReceipt(recipient string, targetSentTimestamps []int64) (err error) {
	call := s.obj.Call("org.asamk.Signal.sendViewedReceipt", 0, recipient, targetSentTimestamps)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.UntrustedIdentity":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Deletes a message with a recipient.
// targetSentTimestamp identifies the message to delete to, recipient is the
// phoneNumber of the chat where to remove from. The returned timestamp can be
// used to identify the corresponding signal reply.
// Might raise `Failure`, `InvalidNumber` exceptions`
func (s *Account) SendRemoteDeleteMessage(targetSentTimestamp int64, recipient string) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendRemoteDeleteMessage", 0, targetSentTimestamp, recipient)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidNumber":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
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

// Deletes a message with multiple recipients.
// targetSentTimestamp identifies the message to delete to, recipients is an
// array of phoneNumbers of the chats where to remove from. The returned
// timestamp can be used to identify the corresponding signal reply.
// Might raise `Failure`, `InvalidNumber` exceptions`
func (s *Account) SendRemoteDeleteMessage_multi(targetSentTimestamp int64, recipients []string) (timestamp int64, err error) {
	call := s.obj.Call("org.asamk.Signal.sendRemoteDeleteMessage", 0, targetSentTimestamp, recipients)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidNumber":
			return 0, fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
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

// Send typing indicator to a recipient.
// recipient is the phoneNumber to send the indicator to, if stop is true the
// typing indicator is removed.
// Might raise `Failure`, `UntrustedIdentity` exceptions.`
func (s *Account) SendTyping(recipient string, stop bool) (err error) {
	call := s.obj.Call("org.asamk.Signal.sendTyping", 0, recipient, stop)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.UntrustedIdentity":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// (Un)block a phoneNumber.
// number is the phoneNumber to block, if block is false the phoneNumber is
// being unblocked. Messages from blocked numbers won't appear on the DBus
// anymore.
// Might raise `InvalidNumber` exception
func (s *Account) SetContactBlocked(number string, block bool) (err error) {
	call := s.obj.Call("org.asamk.Signal.setContactBlocked", 0, number, block)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidNumber":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Set the name correspnding to a phoneNumber in contacts.
// number is the phoneNumber to work on, name is the Name to set in contacts
// (in local storage with signal-cli).
// Might raise `InvalidNumber`, `Failure` exceptions.
func (s *Account) SetContactName(number string, name string) (err error) {
	call := s.obj.Call("org.asamk.Signal.setContactName", 0, number, name)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidNumber":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Delete a contact.
// number is the phoneNumber to delete.
// Might raise `Failure` exception.
func (s *Account) DeleteContact(number string) (err error) {
	call := s.obj.Call("org.asamk.Signal.deleteContact", 0, number)
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

// Delete a recipient.
// number is the phoneNumber.
// Might raise `Failure` exception.
func (s *Account) DeleteRecipient(number string) (err error) {
	call := s.obj.Call("org.asamk.Signal.deleteRecipient", 0, number)
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

// Set expiration timer for a chat.
// number is the phoneNumber of the chat, expiration is the number of seconds
// before messages disappear (set to 0 to disable).
// Might raise `Failure`, `InvalidNumber` exceptions.
func (s *Account) SetExpirationTimer(number string, expiration int32) (err error) {
	call := s.obj.Call("org.asamk.Signal.setExpirationTimer", 0, number, expiration)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.InvalidNumber":
			return fmt.Errorf("signal-cli: %v", err)
		case "org.asamk.Signal.Error.Failure":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Sets a registration lock PIN, to prevent others from registering your
// number.
// pin is the pin to set (resets after 7 days of inactivity).
// Might raise `Failure` exception.
func (s *Account) SetPin(pin string) (err error) {
	// TODO arg pin: can only contain numbers?
	call := s.obj.Call("org.asamk.Signal.setPin", 0, pin)
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

// challenge is the challenge token taken from the proof required error,
// captcha is the token from the solved captcha on the signal website (can be
// used to lift some rate-limits by solving a captcha).
// Might raise `IOErrorException` exception
func (s *Account) SubmitRateLimitChallenge(challenge string, captcha string) (err error) {
	call := s.obj.Call("org.asamk.Signal.submitRateLimitChallenge", 0, challenge, captcha)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.IOErrorException":
			return fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	return nil
}

// Update the own signal profile.
// name is the name to set, about is the about-message, aboutEmoji is the emoji
// for the profile, avatar is the path to an avatar-picture, if remove is true
// the avatar-picture is removed. Strings set to "" result in this property
// being unchanged.
// Might raise: `Failure` exception.
func (s *Account) UpdateProfile(name string, about string, aboutEmoji string, avatar string, remove bool) (err error) {
	call := s.obj.Call("org.asamk.Signal.updateProfile", 0, name, about, aboutEmoji, avatar, remove)
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

// Update the own signal profile.
// givenName and familyName are the different names to set, about is the
// about-message, aboutEmoji is the emoji for the profile, avatar is the path
// to an avatar-picture, if remove is true the avatar-picture is removed.
// Strings set to "" result in this property being unchanged.
// Might raise: `Failure` exception.
func (s *Account) UpdateProfile_firstLastName(givenName string, familyName string, about string, aboutEmoji string, avatar string, remove bool) (err error) {
	call := s.obj.Call("org.asamk.Signal.updateProfile", 0, givenName, familyName, about, aboutEmoji, avatar, remove)
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

// Upload a stickerpack.
// stickerPackPath is the filename to the manifest.json or zip-file in the same
// directory. The returned url is the stickerpack-url after the successful
// upload.
// Might raise `Failure` exception.
func (s *Account) UploadStickerPack(stickerPackPath string) (url string, err error) {
	call := s.obj.Call("org.asamk.Signal.uploadStickerPack", 0, stickerPackPath)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		case "org.asamk.Signal.Error.Failure":
			return "", fmt.Errorf("signal-cli: %v", err)
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&url); err != nil {
		return "", fmt.Errorf("signal-cli: %v", err)
	}
	return url, nil
}

// Get the version of signal-cli
// The returned version is the version-string if signal-cli.
func (s *Account) Version() (version string, err error) {
	call := s.obj.Call("org.asamk.Signal.version", 0)
	if call.Err != nil {
		err := call.Err.(dbus.Error) // panics if assertion does not succeed
		switch err.Name {
		default:
			panic(fmt.Errorf("signal-cli: %v (%s)", err, err.Name))
		}
	}
	if err := call.Store(&version); err != nil {
		return "", fmt.Errorf("signal-cli: %v", err)
	}
	return version, nil
}
