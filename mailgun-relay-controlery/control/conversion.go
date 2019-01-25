package control

import (
	"github.com/Parquery/mailgun-relayery/protoed"
)

// JSONToProto converts a parsed JSON channel to the protobuf channel
// representation.
//
// JSONToProto requires:
// * channel != nil
//
// JSONToProto ensures:
// * protoChan != nil
func JSONToProto(channel *Channel) (protoChan *protoed.Channel) {
	// Pre-condition
	if !(channel != nil) {
		panic("Violated: channel != nil")
	}

	// Post-condition
	defer func() {
		if !(protoChan != nil) {
			panic("Violated: protoChan != nil")
		}
	}()

	sender := jsonToProtoEntity(channel.Sender)
	recipients := jsonToProtoEntityList(channel.Recipients)
	cc := jsonToProtoEntityList(channel.Cc)
	bcc := jsonToProtoEntityList(channel.Bcc)

	protoChan = &protoed.Channel{Descriptor_: string(channel.Descriptor),
		Token: string(channel.Token), Sender: sender,
		Recipients: recipients, Cc: cc, Bcc: bcc, Domain: channel.Domain,
		MinPeriod: channel.MinPeriod, MaxSize: channel.MaxSize}
	return
}

// ProtoToJSON converts a protobuf channel to the parsed JSON channel
// representation.
//
// ProtoToJSON requires:
// * channel != nil
//
// ProtoToJSON ensures:
// * jsonChan != nil
func ProtoToJSON(channel *protoed.Channel) (jsonChan *Channel) {
	// Pre-condition
	if !(channel != nil) {
		panic("Violated: channel != nil")
	}

	// Post-condition
	defer func() {
		if !(jsonChan != nil) {
			panic("Violated: jsonChan != nil")
		}
	}()

	sender := protoToJSONEntity(channel.Sender)
	recipients := protoToJSONEntityList(channel.Recipients)
	cc := protoToJSONEntityList(channel.Cc)
	bcc := protoToJSONEntityList(channel.Bcc)

	return &Channel{Descriptor: Descriptor(channel.Descriptor_),
		Token: Token(channel.Token), Sender: sender,
		Recipients: recipients, Cc: cc, Bcc: bcc, Domain: channel.Domain,
		MinPeriod: channel.MinPeriod, MaxSize: channel.MaxSize}
}

func jsonToProtoEntity(entity Entity) *protoed.Entity {
	name := ""
	if entity.Name != nil {
		name = *entity.Name
	}
	return &protoed.Entity{Name: name, Email: entity.Email}
}

func jsonToProtoEntityList(entities []Entity) []*protoed.Entity {
	protoEntities := []*protoed.Entity{}
	for _, entity := range entities {
		protoEntities = append(protoEntities, jsonToProtoEntity(entity))
	}

	return protoEntities
}

func protoToJSONEntity(entity *protoed.Entity) Entity {
	var name *string
	if entity.Name != "" {
		name = &entity.Name
	}
	return Entity{Name: name, Email: entity.Email}
}

func protoToJSONEntityList(entities []*protoed.Entity) []Entity {
	var jsonEntities []Entity
	for _, entity := range entities {
		jsonEntities = append(jsonEntities, protoToJSONEntity(entity))
	}

	return jsonEntities
}
