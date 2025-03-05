package gpcm

import (
	"fmt"
	"strconv"
	"unicode/utf16"
	"wwfc/common"
	"wwfc/logging"
)

const (
	LangJapanese    = 0x00
	LangEnglish     = 0x01
	LangGerman      = 0x02
	LangFrench      = 0x03
	LangSpanish     = 0x04
	LangItalian     = 0x05
	LangDutch       = 0x06
	LangSimpChinese = 0x07
	LangTradChinese = 0x08
	LangKorean      = 0x09

	LangEnglishEU = 0x81
	LangFrenchEU  = 0x83
	LangSpanishEU = 0x84
)

type WWFCErrorMessage struct {
	ErrorCode  int
	MessageRMC map[byte]string
}

type GPError struct {
	ErrorCode   int
	ErrorString string
	Fatal       bool
	WWFCMessage WWFCErrorMessage
	Reason      string
}

func MakeGPError(errorCode int, errorString string, fatal bool) GPError {
	return GPError{
		ErrorCode:   errorCode,
		ErrorString: errorString,
		Fatal:       fatal,
	}
}

var (
	ErrNone = MakeGPError(0xFFFF, "No error.", false)

	// General errors
	ErrGeneral          = MakeGPError(0x0000, "There was an unknown error.", true)
	ErrParse            = MakeGPError(0x0001, "There was an error parsing an incoming request.", true)
	ErrNotLoggedIn      = MakeGPError(0x0002, "This request cannot be processed because you are not logged in.", true)
	ErrBadSessionKey    = MakeGPError(0x0003, "This request cannot be processed because the session key is invalid.", true)
	ErrDatabase         = MakeGPError(0x0004, "This request cannot be processed because of a database error.", true)
	ErrNetwork          = MakeGPError(0x0005, "There was an error connecting a network socket.", true)
	ErrForcedDisconnect = MakeGPError(0x0006, "This profile has been disconnected by another login.", true)
	ErrConnectionClosed = MakeGPError(0x0007, "The server has closed the connection.", true)
	ErrUDPLayer         = MakeGPError(0x0008, "There was a problem with the UDP layer.", true)

	// Login errors
	ErrLogin                   = MakeGPError(0x0100, "There was an error logging in to the GP backend.", true)
	ErrLoginTimeout            = MakeGPError(0x0101, "The login attempt timed out.", true)
	ErrLoginBadNickname        = MakeGPError(0x0102, "The nickname provided was incorrect.", true)
	ErrLoginBadEmail           = MakeGPError(0x0103, "The email address provided was incorrect.", true)
	ErrLoginBadPassword        = MakeGPError(0x0104, "The password provided is incorrect.", true)
	ErrLoginBadProfile         = MakeGPError(0x0105, "The profile provided was incorrect.", true)
	ErrLoginProfileDeleted     = MakeGPError(0x0106, "The profile has been deleted.", true)
	ErrLoginConnectionFailed   = MakeGPError(0x0107, "The server has refused the connection.", true)
	ErrLoginServerAuthFailed   = MakeGPError(0x0108, "The server could not be authenticated.", true)
	ErrLoginBadUniqueNick      = MakeGPError(0x0109, "The uniquenick provided is incorrect.", true)
	ErrLoginBadPreAuth         = MakeGPError(0x010A, "There was an error validating the pre-authentication.", true)
	ErrLoginLoginTicketInvalid = MakeGPError(0x010B, "The login ticket was unable to be validated.", true)
	ErrLoginLoginTicketExpired = MakeGPError(0x010C, "The login ticket had expired and could not be used.", true)

	// New user errors
	ErrNewUser                  = MakeGPError(0x0200, "There was an error creating a new user.", true)
	ErrNewUserBadNickname       = MakeGPError(0x0201, "A profile with that nick already exists.", true)
	ErrNewUserBadPassword       = MakeGPError(0x0202, "The password does not match the email address.", true)
	ErrNewUserUniqueNickInvalid = MakeGPError(0x0203, "The uniquenick is invalid.", true)
	ErrNewUserUniqueNickInUse   = MakeGPError(0x0204, "The uniquenick is already in use.", true)

	// Update user information errors
	ErrUpdateUserInfo         = MakeGPError(0x0300, "There was an error updating the user information.", false)
	ErrUpdateUserInfoBadEmail = MakeGPError(0x0301, "A user with the email address provided already exists.", false)

	// New profile errors
	ErrNewProfile               = MakeGPError(0x0400, "There was an error creating a new profile.", false)
	ErrNewProfileBadNickname    = MakeGPError(0x0401, "The nickname to be replaced does not exist.", false)
	ErrNewProfileBadOldNickname = MakeGPError(0x0402, "A profile with the nickname provided already exists.", false)

	// Update profile errors
	ErrUpdateProfile            = MakeGPError(0x0500, "There was an error updating the profile information.", false)
	ErrUpdateProfileBadNickname = MakeGPError(0x0501, "A user with the nickname provided already exists.", false)

	// Add friend errors
	ErrAddFriend               = MakeGPError(0x0600, "There was an error adding a buddy.", false)
	ErrAddFriendBadFrom        = MakeGPError(0x0601, "The profile requesting to add a buddy is invalid.", false)
	ErrAddFriendBadNew         = MakeGPError(0x0602, "The profile requested is invalid.", false)
	ErrAddFriendAlreadyFriends = MakeGPError(0x0603, "The profile requested is already a buddy.", false)
	ErrAddFriendLocalBlock     = MakeGPError(0x0604, "The profile requested is on the local profile's block list.", false)
	ErrAddFriendBlocked        = MakeGPError(0x0605, "The profile requested is blocking you.", false)

	// Auth add friend errors
	ErrAuthAdd             = MakeGPError(0x0700, "There was an error authorizing an add buddy request.", false)
	ErrAuthAddBadFrom      = MakeGPError(0x0701, "The profile being authorized is invalid.", false)
	ErrAuthAddBadSignature = MakeGPError(0x0702, "The signature for the authorization is invalid.", false)
	ErrAuthAddLocalBlock   = MakeGPError(0x0703, "The profile requesting authorization is on a block list.", false)
	ErrAuthAddBlocked      = MakeGPError(0x0704, "The profile requested is blocking you.", false)

	// Status errors
	ErrStatus = MakeGPError(0x0800, "There was an error with the status string.", false)

	// Friend message errors
	ErrMessage                    = MakeGPError(0x0900, "There was an error sending a buddy message.", false)
	ErrMessageNotFriends          = MakeGPError(0x0901, "The profile the message was to be sent to is not a buddy.", false)
	ErrMessageExtInfoNotSupported = MakeGPError(0x0902, "The profile does not support extended info keys.", false)
	ErrMessageFriendOffline       = MakeGPError(0x0903, "The buddy to send a message to is offline.", false)

	// Get profile errors
	ErrGetProfile           = MakeGPError(0x0A00, "There was an error getting profile info.", false)
	ErrGetProfileBadProfile = MakeGPError(0x0A01, "The profile info was requested on is invalid.", false)

	// Delete friend errors
	ErrDeleteFriend           = MakeGPError(0x0B00, "There was an error deleting the buddy.", false)
	ErrDeleteFriendNotFriends = MakeGPError(0x0B01, "The buddy to be deleted is not a buddy.", false)

	// Delete profile errors
	ErrDeleteProfile            = MakeGPError(0x0C00, "There was an error deleting the profile.", false)
	ErrDeleteProfileLastProfile = MakeGPError(0x0C01, "The last profile cannot be deleted.", false)

	// Profile searching errors
	ErrSearch              = MakeGPError(0x0D00, "There was an error searching for a profile.", false)
	ErrSearchConnectFailed = MakeGPError(0x0D01, "The search attempt failed to connect to the server.", false)
	ErrSearchTimedOut      = MakeGPError(0x0D02, "The search did not return in a timely fashion.", false)

	// User check errors
	ErrCheck            = MakeGPError(0x0E00, "There was an error checking the user account.", false)
	ErrCheckBadEmail    = MakeGPError(0x0E01, "No account exists with the provided email address.", false)
	ErrCheckBadNickname = MakeGPError(0x0E02, "No such profile exists for the provided email address.", false)
	ErrCheckBadPassword = MakeGPError(0x0E03, "The password is incorrect.", false)

	// Revoke buddy errors
	ErrRevoke           = MakeGPError(0x0F00, "There was an error revoking the buddy.", false)
	ErrRevokeNotFriends = MakeGPError(0x0F01, "You are not a buddy of the profile.", false)

	// Register unique nick errors
	ErrRegisterUniqueNick             = MakeGPError(0x1000, "There was an error registering the uniquenick.", false)
	ErrRegisterUniqueNickTaken        = MakeGPError(0x1001, "The uniquenick is already taken.", false)
	ErrRegisterUniqueNickReserved     = MakeGPError(0x1002, "The uniquenick is reserved.", false)
	ErrRegisterUniqueNickBadNamespace = MakeGPError(0x1003, "Tried to register a nick with no namespace set.", false)

	// Register cdkey errors
	ErrRegisterCDKey             = MakeGPError(0x1100, "There was an error registering the cdkey.", false)
	ErrRegisterCDKeyBadKey       = MakeGPError(0x1101, "The cdkey is invalid.", false)
	ErrRegisterCDKeyAlreadySet   = MakeGPError(0x1102, "The profile has already been registered with a different cdkey.", false)
	ErrRegisterCDKeyAlreadyTaken = MakeGPError(0x1103, "The cdkey has already been registered to another profile.", false)

	// Add to block list errors
	ErrAddBlock               = MakeGPError(0x1200, "There was an error adding the player to the blocked list.", false)
	ErrAddBlockAlreadyBlocked = MakeGPError(0x1201, "The profile specified is already blocked.", false)

	// Remove from block list errors
	ErrRemoveBlock           = MakeGPError(0x1300, "There was an error removing the player from the blocked list.", false)
	ErrRemoveBlockNotBlocked = MakeGPError(0x1301, "The profile specified was not a member of the blocked list.", false)
)

var (
	// Errores de WWFC con mensajes personalizados
	WWFCMsgUnknownLoginError = WWFCErrorMessage{
		ErrorCode: 22000,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Ha ocurrido un error desconocido\n" +
				"al iniciar sesión en PiporGames Network® WFC.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgDolphinSetupRequired = WWFCErrorMessage{
		ErrorCode: 22001,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Se requiere configuración adicional\n" +
				"para usar PiporGames Network® WFC en Dolphin.\n" +
				"Visita pgnweb.ddns.net/legal\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgProfileBannedTOS = WWFCErrorMessage{
		ErrorCode: 22002,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Has sido expulsado de PiporGames Network®\n" +
				"por violar los Términos de Servicio.\n" +
				"Visita pgnweb.ddns.net/legal\n" +
				"\n" +
				"Código de error: %[1]d\n" +
				"Información de soporte: NG%08[2]x",
		},
	}

	WWFCMsgProfileBannedTOSNow = WWFCErrorMessage{
		ErrorCode: 22002,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Has sido expulsado de\n" +
				"PiporGames Network® por violar\n" +
				"los Términos de Servicio.\n" +
				"Visita pgnweb.ddns.net/legal\n" +
				"\n" +
				"Código de error: %[1]d\n" +
				"Información de soporte: NG%08[2]x",
		},
	}

	WWFCMsgProfileRestricted = WWFCErrorMessage{
		ErrorCode: 22003,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"No puedes jugar partidas públicas\n" +
				"debido a una violación de las reglas\n" +
				"de PiporGames Network® WFC.\n" +
				"Visita pgnweb.ddns.net/legal\n" +
				"\n" +
				"Código de error: %[1]d\n" +
				"Información de soporte: NG%08[2]x",
		},
	}

	WWFCMsgProfileRestrictedNow = WWFCErrorMessage{
		ErrorCode: 22003,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Has sido expulsado de las partidas públicas\n" +
				"por violar las reglas de PiporGames Network® WFC.\n" +
				"Visita pgnweb.ddns.net/legal\n" +
				"\n" +
				"Código de error: %[1]d\n" +
				"Información de soporte: NG%08[2]x",
		},
	}

	WWFCMsgProfileRestrictedCustom = WWFCErrorMessage{
		ErrorCode: 22002,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"You are banned from public matches.\n" +
				"Reason: %s" +
				"Error Code: %[1]d\n" +
				"Support Info: NG%08[2]x\n",
		},
	}

	WWFCMsgProfileRestrictedNowCustom = WWFCErrorMessage{
		ErrorCode: 22002,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"You have been banned from public matches.\n" +
				"Reason: %s" +
				"Error Code: %[1]d\n" +
				"Support Info: NG%08[2]x\n",
		},
	}

	WWFCMsgKickedGeneric = WWFCErrorMessage{
		ErrorCode: 22004,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Has sido expulsado de\n" +
				"PiporGames Network® WFC.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgKickedModerator = WWFCErrorMessage{
		ErrorCode: 22004,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Has sido expulsado de\n" +
				"PiporGames Network® WFC por un moderador.\n" +
				"Visita pgnweb.ddns.net/legal\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgKickedRoomHost = WWFCErrorMessage{
		ErrorCode: 22004,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Has sido expulsado de la sala de amigos\n" +
				"por el creador de la sala.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgKickedCustom = WWFCErrorMessage{
		ErrorCode: 22002,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"You have been kicked from WiiLink WFC.\n" +
				"Reason: %s" +
				"Error Code: %[1]d\n" +
				"Support Info: NG%08[2]x\n",
		},
	}

	WWFCMsgConsoleMismatch = WWFCErrorMessage{
		ErrorCode: 22005,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"La consola que estás usando no es\n" +
				"el dispositivo con el que se registró\n" +
				"este perfil.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgConsoleMismatchDolphin = WWFCErrorMessage{
		ErrorCode: 22005,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"La consola que estás usando no es\n" +
				"el dispositivo con el que se registró\n" +
				"este perfil. Asegúrate de haber configurado\n" +
				"correctamente tu NAND.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgProfileIDInvalid = WWFCErrorMessage{
		ErrorCode: 22006,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"El ID de perfil que intentas registrar\n" +
				"no es válido.\n" +
				"Por favor, crea una nueva licencia.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgProfileIDInUse = WWFCErrorMessage{
		ErrorCode: 22007,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"El código de amigo que intentas registrar\n" +
				"ya está en uso.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgPayloadInvalid = WWFCErrorMessage{
		ErrorCode: 22008,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"La carga de datos de PiporGames Network® WFC es inválida.\n" +
				"Intenta reiniciar el juego.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgInvalidELO = WWFCErrorMessage{
		ErrorCode: 22009,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Fuiste desconectado de\n" +
				"PiporGames Network® WFC debido a un\n" +
				"valor de VR o BR inválido.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}

	WWFCMsgTooManyFramesDropped = WWFCErrorMessage{
		ErrorCode: 22010,
		MessageRMC: map[byte]string{
			LangEnglish: "" +
				"Tu juego está perdiendo demasiados fotogramas.\n" +
				"Elimina cualquier modificación que pueda\n" +
				"estar causando problemas de rendimiento para\n" +
				"evitar ser expulsado de PiporGames Network® WFC.\n" +
				"\n" +
				"Código de error: %[1]d",
		},
	}
)


func (err GPError) GetMessage() string {
	command := common.GameSpyCommand{
		Command:      "error",
		CommandValue: "",
		OtherValues: map[string]string{
			"err":    strconv.Itoa(err.ErrorCode),
			"errmsg": err.ErrorString,
		},
	}

	if err.Fatal {
		command.OtherValues["fatal"] = ""
	}

	return common.CreateGameSpyMessage(command)
}

func (err GPError) GetMessageTranslate(gameName string, region byte, lang byte, cfc uint64, ngid uint32) string {
	command := common.GameSpyCommand{
		Command:      "error",
		CommandValue: "",
		OtherValues: map[string]string{
			"err":    strconv.Itoa(err.ErrorCode),
			"errmsg": err.ErrorString,
		},
	}

	if err.Fatal {
		command.OtherValues["fatal"] = ""
	}

	if err.Fatal && err.WWFCMessage.ErrorCode != 0 {
		switch gameName {
		case "mariokartwii":
			reason := err.Reason
			wwfcMessage := err.WWFCMessage

			if reason == "" {
				reason = "None provided."

				engMsg := err.WWFCMessage.MessageRMC[LangEnglish]

				if engMsg == WWFCMsgKickedCustom.MessageRMC[LangEnglish] {
					wwfcMessage = WWFCMsgKickedModerator
				} else if engMsg == WWFCMsgProfileRestrictedCustom.MessageRMC[LangEnglish] {
					wwfcMessage = WWFCMsgProfileRestricted
				} else if engMsg == WWFCMsgProfileRestrictedNowCustom.MessageRMC[LangEnglish] {
					wwfcMessage = WWFCMsgProfileRestrictedNow
				}
			}

			errMsg := wwfcMessage.MessageRMC[lang]
			if errMsg == "" && lang != LangEnglish {
				if lang == LangSpanishEU {
					errMsg = wwfcMessage.MessageRMC[LangSpanish]
				} else if lang == LangFrenchEU {
					errMsg = wwfcMessage.MessageRMC[LangFrench]
				}

				if errMsg == "" {
					errMsg = wwfcMessage.MessageRMC[LangEnglish]
				}
			}

			if errMsg != "" {
				errMsg = fmt.Sprintf(errMsg, wwfcMessage.ErrorCode, ngid, reason)
				errMsgUTF16 := utf16.Encode([]rune(errMsg))
				errMsgByteArray := common.UTF16ToByteArray(errMsgUTF16)
				command.OtherValues["wwfc_errmsg"] = common.Base64DwcEncoding.EncodeToString(errMsgByteArray)
			}

			command.OtherValues["wwfc_err"] = strconv.Itoa(wwfcMessage.ErrorCode)

		}
	}

	return common.CreateGameSpyMessage(command)
}

func (g *GameSpySession) replyError(err GPError) {
	logging.Error(g.ModuleName, "Reply error:", err.ErrorString)
	if !g.LoginInfoSet {
		msg := err.GetMessage()
		// logging.Info(g.ModuleName, "Sending error message:", msg)
		common.SendPacket(ServerName, g.ConnIndex, []byte(msg))
		if err.Fatal {
			common.CloseConnection(ServerName, g.ConnIndex)
		}
		return
	}

	deviceId := g.User.RestrictedDeviceId
	if deviceId == 0 && len(g.User.NgDeviceId) > 0 {
		deviceId = g.User.NgDeviceId[0]
	}

	msg := err.GetMessageTranslate(g.GameName, g.Region, g.Language, g.ConsoleFriendCode, deviceId)
	// logging.Info(g.ModuleName, "Sending error message:", msg)
	common.SendPacket(ServerName, g.ConnIndex, []byte(msg))
	if err.Fatal {
		common.CloseConnection(ServerName, g.ConnIndex)
	}
}
