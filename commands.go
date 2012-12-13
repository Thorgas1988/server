package graval

import (
	"strconv"
	"strings"
)

type ftpCommand interface {
	RequireParam() bool
	RequireAuth() bool
	Execute(*ftpConn, string)
}

type commandMap map[string]ftpCommand

var (
	commands = commandMap{
		"ALLO": commandAllo{},
		"CDUP": commandCdup{},
		"CWD":  commandCwd{},
		"DELE": commandDele{},
		"EPRT": commandEprt{},
		"MODE": commandMode{},
		"NOOP": commandNoop{},
		"PORT": commandPort{},
		"RMD":  commandRmd{},
		"SYST": commandSyst{},
		"TYPE": commandType{},
		"XCUP": commandCdup{},
		"XCWD": commandCwd{},
		"XRMD": commandRmd{},
	}
)

// commandAllo responds to the ALLO FTP command.
//
// This is essentially a ping from the client so we just respond with an
// basic OK message.
type commandAllo struct{}

func (cmd commandAllo) RequireParam() bool {
	return false
}

func (cmd commandAllo) RequireAuth() bool {
	return false
}

func (cmd commandAllo) Execute(conn *ftpConn, param string) {
	conn.writeMessage(202, "Obsolete")
}

// cmdCdup responds to the CDUP FTP command.
//
// Allows the client change their current directory to the parent.
type commandCdup struct{}

func (cmd commandCdup) RequireParam() bool {
	return false
}

func (cmd commandCdup) RequireAuth() bool {
	return false
}

func (cmd commandCdup) Execute(conn *ftpConn, param string) {
	otherCmd := &commandCwd{}
	otherCmd.Execute(conn, "..")
}

// commandCwd responds to the CWD FTP command. It allows the client to change the
// current working directory.
type commandCwd struct{}

func (cmd commandCwd) RequireParam() bool {
	return false
}

func (cmd commandCwd) RequireAuth() bool {
	return false
}

func (cmd commandCwd) Execute(conn *ftpConn, param string) {
	path := conn.buildPath(param)
	if conn.driver.ChangeDir(path) {
		conn.namePrefix = path
		conn.writeMessage(250, "Directory changed to "+path)
	} else {
		conn.writeMessage(550, "Action not taken")
	}
}

// commandDele responds to the DELE FTP command. It allows the client to delete
// a file
type commandDele struct{}

func (cmd commandDele) RequireParam() bool {
	return false
}

func (cmd commandDele) RequireAuth() bool {
	return false
}

func (cmd commandDele) Execute(conn *ftpConn, param string) {
	path := conn.buildPath(param)
	if conn.driver.DeleteFile(path) {
		conn.writeMessage(250, "File deleted")
	} else {
		conn.writeMessage(550, "Action not taken")
	}
}

// commandEprt responds to the EPRT FTP command. It allows the client to
// request an active data socket with more options than the original PORT
// command. It mainly adds ipv6 support.
type commandEprt struct{}

func (cmd commandEprt) RequireParam() bool {
	return false
}

func (cmd commandEprt) RequireAuth() bool {
	return false
}

func (cmd commandEprt) Execute(conn *ftpConn, param string) {
	delim := string(param[0:1])
	parts := strings.Split(param, delim)
	addressFamily, err := strconv.Atoi(parts[1])
	host := parts[2]
	port, err := strconv.Atoi(parts[3])
	if addressFamily != 1 && addressFamily != 2 {
		conn.writeMessage(522, "Network protocol not supported, use (1,2)")
		return
	}
	socket, err := newActiveSocket(host, port)
	if err != nil {
		conn.writeMessage(425, "Data connection failed")
		return
	}
	conn.dataConn = socket
	conn.writeMessage(200, "Connection established ("+strconv.Itoa(port)+")")
}

// cmdMode responds to the MODE FTP command.
//
// the original FTP spec had various options for hosts to negotiate how data
// would be sent over the data socket, In reality these days (S)tream mode
// is all that is used for the mode - data is just streamed down the data
// socket unchanged.
type commandMode struct{}

func (cmd commandMode) RequireParam() bool {
	return false
}

func (cmd commandMode) RequireAuth() bool {
	return false
}

func (cmd commandMode) Execute(conn *ftpConn, param string) {
	if strings.ToUpper(param) == "S" {
		conn.writeMessage(200, "OK")
	} else {
		conn.writeMessage(504, "MODE is an obsolete command")
	}
}

// cmdNoop responds to the NOOP FTP command.
//
// This is essentially a ping from the client so we just respond with an
// basic 200 message.
type commandNoop struct{}

func (cmd commandNoop) RequireParam() bool {
	return false
}

func (cmd commandNoop) RequireAuth() bool {
	return false
}

func (cmd commandNoop) Execute(conn *ftpConn, param string) {
	conn.writeMessage(200, "OK")
}

// commandPort responds to the PORT FTP command.
//
// The client has opened a listening socket for sending out of band data and
// is requesting that we connect to it
type commandPort struct{}

func (cmd commandPort) RequireParam() bool {
	return false
}

func (cmd commandPort) RequireAuth() bool {
	return false
}

func (cmd commandPort) Execute(conn *ftpConn, param string) {
	nums := strings.Split(param, ",")
	portOne, _ := strconv.Atoi(nums[4])
	portTwo, _ := strconv.Atoi(nums[5])
	port := (portOne * 256) + portTwo
	host := nums[0] + "." + nums[1] + "." + nums[2] + "." + nums[3]
	socket, err := newActiveSocket(host, port)
	if err != nil {
		conn.writeMessage(425, "Data connection failed")
		return
	}
	conn.dataConn = socket
	conn.writeMessage(200, "Connection established ("+strconv.Itoa(port)+")")
}

// cmdRmd responds to the RMD FTP command. It allows the client to delete a
// directory.
type commandRmd struct{}

func (cmd commandRmd) RequireParam() bool {
	return false
}

func (cmd commandRmd) RequireAuth() bool {
	return false
}

func (cmd commandRmd) Execute(conn *ftpConn, param string) {
	path := conn.buildPath(param)
	if conn.driver.DeleteDir(path) {
		conn.writeMessage(250, "Directory deleted")
	} else {
		conn.writeMessage(550, "Action not taken")
	}
}

// commandSyst responds to the SYST FTP command by providing a canned response.
type commandSyst struct{}

func (cmd commandSyst) RequireParam() bool {
	return false
}

func (cmd commandSyst) RequireAuth() bool {
	return false
}

func (cmd commandSyst) Execute(conn *ftpConn, param string) {
	conn.writeMessage(215, "UNIX Type: L8")
}

// commandType responds to the TYPE FTP command.
//
//  like the MODE and STRU commands, TYPE dates back to a time when the FTP
//  protocol was more aware of the content of the files it was transferring, and
//  would sometimes be expected to translate things like EOL markers on the fly.
//
//  Valid options were A(SCII), I(mage), E(BCDIC) or LN (for local type). Since
//  we plan to just accept bytes from the client unchanged, I think Image mode is
//  adequate. The RFC requires we accept ASCII mode however, so accept it, but
//  ignore it.
type commandType struct{}

func (cmd commandType) RequireParam() bool {
	return false
}

func (cmd commandType) RequireAuth() bool {
	return false
}

func (cmd commandType) Execute(conn *ftpConn, param string) {
	if strings.ToUpper(param) == "A" {
		conn.writeMessage(200, "Type set to ASCII")
	} else if strings.ToUpper(param) == "I" {
		conn.writeMessage(200, "Type set to binary")
	} else {
		conn.writeMessage(500, "Invalid type")
	}
}