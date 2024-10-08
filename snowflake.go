package snowflake

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

const (
	epoch             = int64(1577808000000)                           // Set start time (timestamp / millisecond): 2020-01-01 00, valid for 69 years
	timestampBits     = uint(41)                                       // Number of digits occupied by timestamp
	datacenterIDBits  = uint(5)                                        // Number of places occupied by id in the data center
	workerIDBits      = uint(5)                                        // Number of bits occupied by machine id
	sequenceBits      = uint(12)                                       // The number of digits occupied by the sequence
	timestampMax      = int64(-1 ^ (-1 << timestampBits))              // Timestamp maximum
	datacenterIDMax   = int64(-1 ^ (-1 << datacenterIDBits))           // Maximum number of data center id supported.
	workerIDMax       = int64(-1 ^ (-1 << workerIDBits))               // Maximum number of machine id supported
	sequenceMask      = int64(-1 ^ (-1 << sequenceBits))               // Maximum number of sequence id supported
	workerIDShift     = sequenceBits                                   // Number of left shifts of machine id
	datacenterIDShift = sequenceBits + workerIDBits                    // Data Center id left shift
	timestampShift    = sequenceBits + workerIDBits + datacenterIDBits // Timestamp left shift
)

// An ID is a custom type used for a snowflake ID.  This is used so we can
// attach methods onto the ID.
type ID int64

const encodeBase32Map = "ybndrfg8ejkmcpqxot1uwisza345h769"

var decodeBase32Map [256]byte

const encodeBase58Map = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"

var decodeBase58Map [256]byte

// A JSONSyntaxError is returned from UnmarshalJSON if an invalid ID is provided.
type JSONSyntaxError struct{ original []byte }

func (j JSONSyntaxError) Error() string {
	return fmt.Sprintf("invalid snowflake ID %q", string(j.original))
}

// ErrInvalidBase58 is returned by ParseBase58 when given an invalid []byte
var ErrInvalidBase58 = errors.New("invalid base58")

// ErrInvalidBase32 is returned by ParseBase32 when given an invalid []byte
var ErrInvalidBase32 = errors.New("invalid base32")

// Create maps for decoding Base58/Base32.
// This speeds up the process tremendously.
func init() {

	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase58Map); i++ {
		decodeBase58Map[encodeBase58Map[i]] = byte(i)
	}

	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase32Map); i++ {
		decodeBase32Map[encodeBase32Map[i]] = byte(i)
	}
}

// Snowflake is a custom type
type Snowflake struct {
	sync.Mutex
	timestamp    int64
	workerID     int64
	datacenterID int64
	sequence     int64
}

// NewSnowflake returns a new snowflake node that can be used to generate snowflake
func NewSnowflake(datacenterID, workerID int64) (*Snowflake, error) {
	if datacenterID < 0 || datacenterID > datacenterIDMax {
		return nil, fmt.Errorf("datacenterid must be between 0 and %d", datacenterIDMax-1)
	}
	if workerID < 0 || workerID > workerIDMax {
		return nil, fmt.Errorf("workerid must be between 0 and %d", workerIDMax-1)
	}
	return &Snowflake{
		timestamp:    0,
		datacenterID: datacenterID,
		workerID:     workerID,
		sequence:     0,
	}, nil
}

// NextVal creates and returns a unique snowflake ID
// To help guarantee uniqueness
// - Make sure your system is keeping accurate system time
// - Make sure you never have multiple nodes running with the same node ID
func (s *Snowflake) NextVal() ID {
	s.Lock()
	now := time.Now().UnixNano() / 1000000 // 转毫秒
	if s.timestamp == now {
		// When id is generated multiple times under the same timestamp (precision: millisecond), the sequence number will be increased.
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// If the current sequence exceeds the 12bit length, you need to wait for the next millisecond
			// Sequence:0 will be used in the next millisecond
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		// Use the serial number directly under different timestamps (precision: milliseconds): 0
		s.sequence = 0
	}
	t := now - epoch
	if t > timestampMax {
		s.Unlock()
		fmt.Printf("epoch must be between 0 and %d", timestampMax-1)
		return 0
	}
	s.timestamp = now
	r := ID((t)<<timestampShift | (s.datacenterID << datacenterIDShift) | (s.workerID << workerIDShift) | (s.sequence))
	s.Unlock()
	return r
}

// GetDeviceID returns an int64 of the snowflake center ID and machine ID number
func GetDeviceID(sid int64) (datacenterID, workerID int64) {
	datacenterID = (sid >> datacenterIDShift) & datacenterIDMax
	workerID = (sid >> workerIDShift) & workerIDMax
	return
}

// GetTimestamp returns an int64 unix timestamp in milliseconds of the snowflake ID time
func GetTimestamp(sid ID) (timestamp int64) {
	timestamp = (int64(sid) >> timestampShift) & timestampMax
	return
}

// GetGenTimestamp returns Get the timestamp when the ID was created
func GetGenTimestamp(sid ID) (timestamp int64) {
	timestamp = GetTimestamp(sid) + epoch
	return
}

// GetGenTime returns Gets the time string when the ID was created (precision: seconds)
func GetGenTime(sid ID) (t string) {
	// The timestamp / 1000 obtained by GetGenTimestamp needs to be converted into seconds
	t = time.Unix(GetGenTimestamp(sid)/1000, 0).Format("2006-01-02 15:04:05")
	return
}

// GetTimestampStatus returns an float64 unix timestamp in milliseconds of the snowflake ID time Get the percentage of timestamps used: range (0.0-1.0)
func GetTimestampStatus() (state float64) {
	state = float64(time.Now().UnixNano()/1000000-epoch) / float64(timestampMax)
	return
}

// Int64 returns an int64 of the snowflake ID
func (sid ID) Int64() int64 {
	return int64(sid)
}

// ParseInt64 converts an int64 into a snowflake ID
func ParseInt64(id int64) ID {
	return ID(id)
}

// String returns a string of the snowflake ID
func (sid ID) String() string {
	return strconv.FormatInt(int64(sid), 10)
}

// ParseString converts a string into a snowflake ID
func ParseString(sid string) (ID, error) {
	i, err := strconv.ParseInt(sid, 10, 64)
	return ID(i), err

}

// Base2 returns a string base2 of the snowflake ID
func (sid ID) Base2() string {
	return strconv.FormatInt(int64(sid), 2)
}

// ParseBase2 converts a Base2 string into a snowflake ID
func ParseBase2(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 2, 64)
	return ID(i), err
}

// Base32 uses the z-base-32 character set but encodes and decodes similar
// to base58, allowing it to create an even smaller result string.
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func (sid ID) Base32() string {

	if sid < 32 {
		return string(encodeBase32Map[sid])
	}

	b := make([]byte, 0, 12)
	for sid >= 32 {
		b = append(b, encodeBase32Map[sid%32])
		sid /= 32
	}
	b = append(b, encodeBase32Map[sid])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// ParseBase32 parses a base32 []byte into a snowflake ID
// NOTE: There are many different base32 implementations so becareful when
// doing any interoperation.
func ParseBase32(b []byte) (ID, error) {

	var id int64

	for i := range b {
		if decodeBase32Map[b[i]] == 0xFF {
			return -1, ErrInvalidBase32
		}
		id = id*32 + int64(decodeBase32Map[b[i]])
	}

	return ID(id), nil
}

// Base36 returns a base36 string of the snowflake ID
func (sid ID) Base36() string {
	return strconv.FormatInt(int64(sid), 36)
}

// ParseBase36 converts a Base36 string into a snowflake ID
func ParseBase36(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 36, 64)
	return ID(i), err
}

// Base58 returns a base58 string of the snowflake ID
func (sid ID) Base58() string {

	if sid < 58 {
		return string(encodeBase58Map[sid])
	}

	b := make([]byte, 0, 11)
	for sid >= 58 {
		b = append(b, encodeBase58Map[sid%58])
		sid /= 58
	}
	b = append(b, encodeBase58Map[sid])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

// ParseBase58 parses a base58 []byte into a snowflake ID
func ParseBase58(b []byte) (ID, error) {

	var id int64

	for i := range b {
		if decodeBase58Map[b[i]] == 0xFF {
			return -1, ErrInvalidBase58
		}
		id = id*58 + int64(decodeBase58Map[b[i]])
	}

	return ID(id), nil
}

// Base64 returns a base64 string of the snowflake ID
func (sid ID) Base64() string {
	return base64.StdEncoding.EncodeToString(sid.Bytes())
}

// ParseBase64 converts a base64 string into a snowflake ID
func ParseBase64(id string) (ID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return -1, err
	}
	return ParseBytes(b)

}

// Bytes return a byte slice of the snowflake ID
func (sid ID) Bytes() []byte {
	return []byte(sid.String())
}

// ParseBytes converts a byte slice into a snowflake ID
func ParseBytes(id []byte) (ID, error) {
	i, err := strconv.ParseInt(string(id), 10, 64)
	return ID(i), err
}

// IntBytes returns an array of bytes of the snowflake ID, encoded as a
// big endian integer.
func (sid ID) IntBytes() [8]byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(sid))
	return b
}

// ParseIntBytes converts an array of bytes encoded as big endian integer as
// a snowflake ID
func ParseIntBytes(id [8]byte) ID {
	return ID(int64(binary.BigEndian.Uint64(id[:])))
}

// MarshalJSON returns a json byte array string of the snowflake ID.
func (sid ID) MarshalJSON() ([]byte, error) {
	buff := make([]byte, 0, 22)
	buff = append(buff, '"')
	buff = strconv.AppendInt(buff, int64(sid), 10)
	buff = append(buff, '"')
	return buff, nil
}

// UnmarshalJSON converts a json byte array of a snowflake ID into an ID type.
func (sid *ID) UnmarshalJSON(b []byte) error {
	if len(b) < 3 || b[0] != '"' || b[len(b)-1] != '"' {
		return JSONSyntaxError{b}
	}

	i, err := strconv.ParseInt(string(b[1:len(b)-1]), 10, 64)
	if err != nil {
		return err
	}

	*sid = ID(i)
	return nil
}

// Time returns an int64 unix timestamp in milliseconds of the snowflake ID time
func (sid ID) Time() int64 {
	return (int64(sid) >> timestampShift) + epoch
}

// GetTimestampMax Timestamp maximum
func GetTimestampMax() int64 {
	return timestampMax
}

// GetDatacenterIDMax Maximum number of data center id supported.
func GetDatacenterIDMax() int64 {
	return datacenterIDMax
}

// GetWorkerIDMax Maximum number of machine id supported
func GetWorkerIDMax() int64 {
	return workerIDMax
}

// GetSequenceMask Maximum number of sequence id supported
func GetSequenceMask() int64 {
	return sequenceMask
}
