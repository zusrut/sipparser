package sipparser

import (
	"errors"
	"fmt"
	"strings"
)

const (
	sipParseStateStartLine    = "SipParseStateStartLine"
	sipParseStateCrlf         = "SipMsgStateCrlf"
	sipParseStateBody         = "SipMsgStateBody"
	sipParseStateHeaders      = "SipMsgStateHeaders"
	sipParseStateParseHeaders = "SipMsgStateParseHeaders"
	uintCR                    = '\r'
	uintLF                    = '\n'
	CR                        = "\r"
	LF                        = "\n"
	CALLING_PARTY_DEFAULT     = "default"
	CALLING_PARTY_RPID        = "rpid"
	CALLING_PARTY_PAID        = "paid"
)

type CallingPartyInfo struct {
	Name      string
	Number    string
	Anonymous bool
}

type Header struct {
	Header string
	Val    string
}

func (h *Header) String() string {
	return fmt.Sprintf("%s: %s", h.Header, h.Val)
}

type sipParserStateFn func(s *SipMsg) sipParserStateFn

type SipMsg struct {
	State            string
	Error            error
	Msg              string
	CallingParty     *CallingPartyInfo
	Body             string
	StartLine        *StartLine
	Headers          []*Header
	Authorization    *Authorization
	AuthVal          string
	AuthUser         string
	ContentLength    string
	ContentType      string
	From             *From
	FromUser         string
	FromHost         string
	FromTag          string
	MaxForwards      string
	Organization     string
	To               *From
	ToUser           string
	ToHost           string
	ToTag            string
	Contact          *From
	ContactVal       string
	ContactUser      string
	ContactHost      string
	ContactPort      int
	CallID           string
	XCallID          string
	Cseq             *Cseq
	CseqMethod       string
	CseqVal          string
	Reason           *Reason
	ReasonVal        string
	RTPStatVal       string
	Via              []*Via
	ViaOne           string
	ViaOneBranch     string
	Privacy          string
	RemotePartyIdVal string
	DiversionVal     string
	RemotePartyId    *RemotePartyId
	PAssertedIdVal   string
	PaiUser          string
	PaiHost          string
	PAssertedId      *PAssertedId
	UserAgent        string
	Server           string
	eof              int
	hdr              string
	hdrv             string

	//Accept             *Accept
	//AlertInfo          string
	//Allow              []string
	//AllowEvents        []string
	//ContentDisposition *ContentDisposition
	//ContentLengthInt   int
	//MaxForwardsInt     int
	//ProxyAuthenticate  *Authorization
	//ProxyRequire       []string
	//Rack               *Rack
	//Rseq               string
	//RseqInt            int
	//RecordRoute        []*URI
	//RTPStat            *RTPStat
	//Route              []*URI
	//Require            []string
	//Unsupported        []string
	//Subject            string
	//Supported          []string
	//Warning            *Warning
	//WWWAuthenticate    *Authorization
}

func (s *SipMsg) run() {
	for state := parseSip; state != nil; {
		state = state(s)
	}
}

func (s *SipMsg) addError(err string) sipParserStateFn {
	s.Error = errors.New(err)
	return nil
}

func (s *SipMsg) addErrorNoReturn(err string) {
	s.Error = errors.New(err)
}

func (s *SipMsg) addHdr(str string) {
	if str == "" || str == " " {
		return
	}
	sp := strings.IndexRune(str, ':')
	if sp == -1 {
		return
	}
	//s.hdr = strings.ToLower(strings.TrimSpace(str[0:sp]))

	s.hdr = str[0:sp]
	if len(str[0:sp]) > 1 {
		if str[sp-1] == '\t' || str[sp-1] == ' ' {
			s.hdr = str[0 : sp-1]
		}
	}
	if len(str[0:sp]) > 2 {
		if str[sp-2:sp] == "  " {
			s.hdr = str[0 : sp-2]
		}
	}
	if len(str[0:sp]) > 3 {
		if str[sp-3:sp] == "   " {
			s.hdr = str[0 : sp-3]
		}
	}
	if len(str[0:sp]) > 4 {
		if str[sp-4:sp] == "    " {
			s.hdr = str[0 : sp-4]
		}
	}
	if len(str)-1 >= sp+1 {
		s.hdrv = cleanWs(str[sp+1:])
	} else {
		s.Error = fmt.Errorf("addHdr err: no valid header: %s", s.hdr)
		s.hdrv = ""
	}
	switch {
	case s.hdr == SIP_HDR_ACCEPT || s.hdr == SIP_HDR_ACCEPT_ENCODING || s.hdr == SIP_HDR_ACCEPT_LANGUAGE:
	//s.parseAccept(s.hdrv)
	case s.hdr == SIP_HDR_ALLOW:
	//s.parseAllow(s.hdrv)
	case s.hdr == SIP_HDR_ALLOW_EVENTS || s.hdr == SIP_HDR_ALLOW_EVENTS_CMP:
	//s.parseAllowEvents(s.hdrv)
	//case s.hdr == SIP_HDR_AUTHORIZATION || s.hdr == SIP_HDR_PROXY_AUTHORIZATION:
	case s.hdr == "Authorization" || s.hdr == "authorization" || s.hdr == "Proxy-Authorization" || s.hdr == "proxy-authorization":
		s.parseAuthorization(s.hdrv)
	//case s.hdr == SIP_HDR_CALL_ID || s.hdr == SIP_HDR_CALL_ID_CMP:
	case s.hdr == "Call-ID" || s.hdr == "Call-Id" || s.hdr == "Call-id" || s.hdr == "call-id" || s.hdr == "I" || s.hdr == "i":
		s.CallID = s.hdrv
	//case s.hdr == HOMER_HDR_X_CID:
	case s.hdr == "P-Charging-Vector":
		s.XCallID = s.hdrv
	case s.hdr == "X-CID" || s.hdr == "x-cid" || s.hdr == "XCall-ID" || s.hdr == "XCall-Id":
		s.XCallID = s.hdrv
	//case s.hdr == SIP_HDR_CONTACT || s.hdr == SIP_HDR_CONTACT_CMP:
	case s.hdr == "Contact" || s.hdr == "contact" || s.hdr == "M" || s.hdr == "m":
		s.ContactVal = s.hdrv
		s.parseContact(str)
	case s.hdr == SIP_HDR_CONTENT_DISPOSITION:
	//s.parseContentDisposition(s.hdrv)
	//case s.hdr == SIP_HDR_CONTENT_LENGTH || s.hdr == SIP_HDR_CONTENT_LENGTH_CMP:
	case s.hdr == "Content-Length" || s.hdr == "content-length" || s.hdr == "L" || s.hdr == "l":
		s.ContentLength = s.hdrv
	//case s.hdr == SIP_HDR_CONTENT_TYPE || s.hdr == SIP_HDR_CONTENT_TYPE_CMP:
	case s.hdr == "Content-Type" || s.hdr == "content-type" || s.hdr == "C" || s.hdr == "c":
		s.ContentType = s.hdrv
	//case s.hdr == SIP_HDR_CSEQ:
	case s.hdr == "CSeq" || s.hdr == "Cseq" || s.hdr == "cSeq" || s.hdr == "cseq":
		s.CseqVal = s.hdrv
		s.parseCseq(s.hdrv)
	//case s.hdr == SIP_HDR_FROM || s.hdr == SIP_HDR_FROM_CMP:
	case s.hdr == "From" || s.hdr == "from" || s.hdr == "F" || s.hdr == "f":
		s.parseFrom(s.hdrv)
	//case s.hdr == SIP_HDR_MAX_FORWARDS:
	case s.hdr == "Max-Forwards" || s.hdr == "max-forwards":
		s.MaxForwards = s.hdrv
	//case s.hdr == SIP_HDR_ORGANIZATION:
	case s.hdr == "Organization" || s.hdr == "organization":
		s.Organization = s.hdrv
	//case s.hdr == SIP_HDR_P_ASSERTED_IDENTITY:
	case s.hdr == "P-Asserted-Identity" || s.hdr == "p-asserted-identity":
		s.PAssertedIdVal = s.hdrv
		s.parsePAssertedId(s.hdrv)
	//case s.hdr == SIP_HDR_PRIVACY:
	case s.hdr == "Privacy" || s.hdr == "privacy":
		s.Privacy = s.hdrv
	//case s.hdr == SIP_HDR_PROXY_AUTHENTICATE:
	case s.hdr == "Proxy-Authenticate" || s.hdr == "proxy-authenticate":
	//s.parseProxyAuthenticate(s.hdrv)
	case s.hdr == SIP_HDR_RACK:
	//s.parseRack(s.hdrv)
	//case s.hdr == SIP_HDR_REASON:
	case s.hdr == "Reason" || s.hdr == "reason":
		s.ReasonVal = s.hdrv
		//s.parseReason(s.hdrv)
	case s.hdr == SIP_HDR_RECORD_ROUTE:
	//s.parseRecordRoute(s.hdrv)
	//case s.hdr == SIP_HDR_REMOTE_PARTY_ID:
	case s.hdr == "Remote-Party-Id" || s.hdr == "remote-party-id":
		s.RemotePartyIdVal = s.hdrv
	//case s.hdr == SIP_HDR_DIVERSION:
	case s.hdr == "Diversion" || s.hdr == "diversion":
		s.DiversionVal = s.hdrv
	case s.hdr == SIP_HDR_ROUTE:
	//s.parseRoute(s.hdrv)
	//case s.hdr == SIP_HDR_X_RTP_STAT:
	case s.hdr == "X-RTP-Stat" || s.hdr == "x-rtp-stat" || s.hdr == "X-Rtp-Stat":
		s.parseRTPStat(s.hdrv)
	//case s.hdr == SIP_HDR_SERVER:
	case s.hdr == "Server" || s.hdr == "server":
		s.Server = s.hdrv
	case s.hdr == SIP_HDR_SUPPORTED:
	//s.parseSupported(s.hdrv)
	//case s.hdr == SIP_HDR_TO || s.hdr == SIP_HDR_TO_CMP:
	case s.hdr == "To" || s.hdr == "TO" || s.hdr == "to" || s.hdr == "T" || s.hdr == "t":
		s.parseTo(s.hdrv)
	case s.hdr == SIP_HDR_UNSUPPORTED:
	//s.parseUnsupported(s.hdrv)
	//case s.hdr == SIP_HDR_USER_AGENT:
	case s.hdr == "User-Agent" || s.hdr == "user-agent":
		s.UserAgent = s.hdrv
	//case s.hdr == SIP_HDR_VIA || s.hdr == SIP_HDR_VIA_CMP:
	case s.hdr == "Via" || s.hdr == "via" || s.hdr == "V" || s.hdr == "v":
		s.parseVia(s.hdrv)
	case s.hdr == SIP_HDR_WARNING:
	//s.parseWarning(s.hdrv)
	case s.hdr == SIP_HDR_WWW_AUTHENTICATE:
	//s.parseWWWAuthenticate(s.hdrv)
	default:
		/* 		// Append unkown headers to s.Headers
		   		if s.Headers == nil {
		   			s.Headers = make([]*Header, 0)
		   		}
		   		s.Headers = append(s.Headers, &Header{s.hdr, s.hdrv}) */
	}
}

func (s *SipMsg) GetRURIParamBool(str string) bool {
	if s.StartLine == nil || s.StartLine.URI == nil {
		return false
	}
	for i := range s.StartLine.URI.UriParams {
		if s.StartLine.URI.UriParams[i].Param == str {
			return true
		}
	}
	return false
}

func (s *SipMsg) GetRURIParamVal(str string) string {
	if s.StartLine == nil || s.StartLine.URI == nil {
		return ""
	}
	for i := range s.StartLine.URI.UriParams {
		if s.StartLine.URI.UriParams[i].Param == str {
			return s.StartLine.URI.UriParams[i].Val
		}
	}
	return ""
}

func (s *SipMsg) GetCallingParty(str string) error {
	switch {
	case str == CALLING_PARTY_RPID:
		return s.getCallingPartyRpid()
	case str == CALLING_PARTY_PAID:
		return s.getCallingPartyPaid()
	}
	return s.getCallingPartyDefault()
}

func (s *SipMsg) getCallingPartyDefault() error {
	if s.From == nil {
		return errors.New("getCallingPartyDefault err: no from header found")
	}
	if s.From.URI == nil {
		return errors.New("getCallingPartyDefault err: no uri found in from header")
	}
	s.CallingParty = &CallingPartyInfo{Name: s.From.Name, Number: s.From.URI.User}
	return nil
}

func (s *SipMsg) getCallingPartyPaid() error {
	if s.PAssertedId == nil {
		if s.PAssertedIdVal == "" {
			return s.getCallingPartyDefault()
		}
		s.parsePAssertedId(s.PAssertedIdVal)
		if s.Error != nil {
			return s.Error
		}
		if s.PAssertedId.URI == nil {
			return errors.New("getCallingPartyPaid err: p-asserted-id uri is nil")
		}
		s.CallingParty = &CallingPartyInfo{Name: s.PAssertedId.Name, Number: s.PAssertedId.URI.User}
		return nil
	}
	if s.PAssertedId.URI == nil {
		return errors.New("getCallingPartyPaid err: p-asserted-id uri is nil")
	}
	s.CallingParty = &CallingPartyInfo{Name: s.PAssertedId.Name, Number: s.PAssertedId.URI.User}
	return nil
}

func (s *SipMsg) getCallingPartyRpid() error {
	if s.RemotePartyId == nil {
		if s.RemotePartyIdVal == "" {
			return s.getCallingPartyDefault()
		}
		s.parseRemotePartyId(s.RemotePartyIdVal)
		if s.Error != nil {
			return s.Error
		}
		if s.RemotePartyId.URI == nil {
			return errors.New("getCallingPartyRpid err: remote party id uri is nil")
		}
		s.CallingParty = &CallingPartyInfo{Name: s.RemotePartyId.Name, Number: s.RemotePartyId.URI.User}
		return nil
	}
	if s.RemotePartyId.URI == nil {
		return errors.New("getCallingPartyRpid err: remote party id uri is nil")
	}
	s.CallingParty = &CallingPartyInfo{Name: s.RemotePartyId.Name, Number: s.RemotePartyId.URI.User}
	return nil
}

/* func (s *SipMsg) parseAccept(str string) {
	s.Accept = &Accept{Val: str}
	s.Accept.parse()
} */

/* func (s *SipMsg) parseAllow(str string) {
	s.Allow = getCommaSeperated(str)
	if s.Allow == nil {
		s.Allow = []string{str}
	}
} */

/* func (s *SipMsg) parseAllowEvents(str string) {
	s.AllowEvents = getCommaSeperated(str)
	if s.AllowEvents == nil {
		s.Allow = []string{str}
	}
} */

func (s *SipMsg) parseAuthorization(str string) {
	s.Authorization = &Authorization{Val: str}
	if s.Error = s.Authorization.parse(); s.Error == nil {
		s.AuthUser = s.Authorization.Username
		s.AuthVal = s.Authorization.Val
	}
}

func (s *SipMsg) parseContact(str string) {
	s.Contact = getFrom(str)
	if s.Contact.Error == nil {
		s.ContactUser = s.Contact.URI.User
		s.ContactHost = s.Contact.URI.Host
		s.ContactPort = s.Contact.URI.PortInt
	} else {
		s.Error = s.Contact.Error
	}
}

func (s *SipMsg) ParseContact(str string) {
	s.parseContact(str)
}

/* func (s *SipMsg) parseContentDisposition(str string) {
	s.ContentDisposition = &ContentDisposition{Val: str}
	s.ContentDisposition.parse()
} */

func (s *SipMsg) parseCseq(str string) {
	s.Cseq = &Cseq{Val: str}
	if s.Error = s.Cseq.parse(); s.Error == nil {
		s.CseqMethod = s.Cseq.Method
	}
}

func (s *SipMsg) parseFrom(str string) {
	s.From = getFrom(str)
	if s.From.Error == nil {
		s.FromUser = s.From.URI.User
		s.FromHost = s.From.URI.Host
		s.FromTag = s.From.Tag
	} else {
		s.Error = s.From.Error
	}
}

func (s *SipMsg) parsePAssertedId(str string) {
	s.PAssertedId = &PAssertedId{Val: str}
	s.PAssertedId.parse()
	if s.PAssertedId.Error == nil {
		if s.PaiUser == "" {
			s.PaiUser = s.PAssertedId.URI.User
		}
		if s.PaiHost == "" {
			s.PaiHost = s.PAssertedId.URI.Host
		}
	} else {
		s.PaiUser = s.PAssertedIdVal
	}
}

func (s *SipMsg) ParsePAssertedId(str string) {
	s.parsePAssertedId(str)
}

/* func (s *SipMsg) parseProxyAuthenticate(str string) {
	s.ProxyAuthenticate = &Authorization{Val: str}
	s.Error = s.ProxyAuthenticate.parse()
} */

/* func (s *SipMsg) parseRack(str string) {
	s.Rack = &Rack{Val: str}
	s.Error = s.Rack.parse()
} */

func (s *SipMsg) parseReason(str string) {
	s.Reason = &Reason{Val: str}
	s.Reason.parse()
}

func (s *SipMsg) parseRTPStat(str string) {
	//s.RTPStat = &RTPStat{Val: str}
	//s.RTPStat.parse()
	s.RTPStatVal = str
}

/* func (s *SipMsg) parseRecordRoute(str string) {
	cs := []string{str}
	for rt := range cs {
		left := 0
		right := 0
		for i := range cs[rt] {
			if cs[rt][i] == '<' && left == 0 {
				left = i
			}
			if cs[rt][i] == '>' && right == 0 {
				right = i
			}
		}
		if left < right {
			u := ParseURI(cs[rt][left+1 : right])
			if u.Error != nil {
				s.Error = fmt.Errorf("parseRecordRoute err: received err parsing uri: %v", u.Error)
				return
			}
			if s.RecordRoute == nil {
				s.RecordRoute = []*URI{u}
			}
			s.RecordRoute = append(s.RecordRoute, u)
		}
	}
	return
} */

func (s *SipMsg) parseRemotePartyId(str string) {
	s.RemotePartyId = &RemotePartyId{Val: str}
	s.RemotePartyId.parse()
	if s.RemotePartyId.Error != nil {
		s.Error = s.RemotePartyId.Error
	}
}

func (s *SipMsg) ParseRemotePartyId(str string) {
	s.parseRemotePartyId(str)
}

/* func (s *SipMsg) parseRequire(str string) {
	s.Require = getCommaSeperated(str)
	if s.Require == nil {
		s.Require = []string{str}
	}
} */

/* func (s *SipMsg) parseRoute(str string) {
	cs := getCommaSeperated(str)
	for rt := range cs {
		left := 0
		right := 0
		for i := range cs[rt] {
			if cs[rt][i] == '<' && left == 0 {
				left = i
			}
			if cs[rt][i] == '>' && right == 0 {
				right = i
			}
		}
		if left < right {
			u := ParseURI(cs[rt][left+1 : right])
			if u.Error != nil {
				s.Error = fmt.Errorf("parseRoute err: received err parsing uri: %v", u.Error)
				return
			}
			if s.Route == nil {
				s.Route = []*URI{u}
			}
			s.Route = append(s.Route, u)
		}
	}
} */

func (s *SipMsg) parseStartLine(str string) {
	s.State = sipParseStateStartLine
	s.StartLine = ParseStartLine(str)
	if s.StartLine.Error != nil {
		s.Error = fmt.Errorf("parseStartLine err: received err while parsing start line: %v", s.StartLine.Error)
	}
}

/* func (s *SipMsg) parseSupported(str string) {
	s.Supported = getCommaSeperated(str)
	if s.Supported == nil {
		s.Supported = []string{str}
	}
} */

func (s *SipMsg) parseTo(str string) {
	s.To = getFrom(str)
	if s.To.Error == nil {
		s.ToUser = s.To.URI.User
		s.ToHost = s.To.URI.Host
		s.ToTag = s.To.Tag
	} else {
		s.Error = s.To.Error
	}
}

/* func (s *SipMsg) parseUnsupported(str string) {
	s.Unsupported = getCommaSeperated(str)
	if s.Unsupported == nil {
		s.Unsupported = []string{str}
	}
} */

func (s *SipMsg) parseVia(str string) {
	vs := &vias{via: str}
	vs.parse()
	if vs.err != nil {
		s.Error = vs.err
		return
	}
	s.ViaOne = vs.vias[0].Via
	s.ViaOneBranch = vs.vias[0].Branch
	for _, v := range vs.vias {
		s.Via = append(s.Via, v)
	}
}

/* func (s *SipMsg) parseWarning(str string) {
	s.Warning = &Warning{Val: str}
	s.Error = s.Warning.parse()
} */

/* func (s *SipMsg) parseWWWAuthenticate(str string) {
	s.WWWAuthenticate = &Authorization{Val: str}
	s.Error = s.WWWAuthenticate.parse()
} */

func getHeaders(s *SipMsg) sipParserStateFn {
	s.State = sipParseStateHeaders
	var hdr string

	for curPos, crlfPos := 0, 0; curPos < s.eof+2 && s.eof+2 <= len(s.Msg); curPos += 2 {
		crlfPos = strings.Index(s.Msg[curPos:s.eof+2], "\r\n")
		hdr = s.Msg[curPos : curPos+crlfPos]

		if len(hdr) > 1 {
			if hdr[0] == '\t' || hdr[0] == ' ' {
				hdr = hdr[1:]
			}
		}
		if len(hdr) > 2 {
			if hdr[0:2] == "  " {
				hdr = hdr[2:]
			}
		}
		if len(hdr) > 3 {
			if hdr[0:3] == "   " {
				hdr = hdr[3:]
			}
		}
		if len(hdr) > 4 {
			if hdr[0:4] == "    " {
				hdr = hdr[4:]
			}
		}
		if curPos == 0 {
			s.parseStartLine(hdr)
		} else {
			s.addHdr(hdr)
		}
		if s.Error != nil {
			return nil
		}
		curPos += crlfPos
	}
	return nil
}

func getBody(s *SipMsg) sipParserStateFn {
	s.State = sipParseStateBody
	if len(s.Msg)-1 > s.eof+4 {
		s.Body = s.Msg[s.eof+4:]
	}
	return getHeaders
}

func ParseMsg(str string) (s *SipMsg) {
	headersEnd := strings.Index(str, "\r\n\r\n")
	if headersEnd == -1 {
		headersEnd = strings.LastIndex(str, "\r\n")
	}
	s = &SipMsg{Msg: str, eof: headersEnd}
	if s.eof == -1 {
		s.Error = errors.New("ParseMsg: err parsing msg no SIP eof found")
		return s
	}
	s.run()
	return s
}

func parseSip(s *SipMsg) sipParserStateFn {
	if s.Error != nil {
		return nil
	}
	return getBody
}
