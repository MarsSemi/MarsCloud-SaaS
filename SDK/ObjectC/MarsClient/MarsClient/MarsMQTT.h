#ifndef __MARSMQTT__
#define __MARSMQTT__
//-----------------------------------------------------------
#pragma once
//-----------------------------------------------------------
#import "MarsClient.h"
#import "MQTTClient.h"
//-----------------------------------------------------------
@interface MarsMQTT : NSObject
{
    MarsClient *_MarsClient;
    MQTTSession *_MQTTClient;
}
//-----------------------------------------------------------
- (void)connected :(MQTTSession *)_session;
- (void)connectionClosed :(MQTTSession *)_session;
- (void)newMessage:(MQTTSession *)_session data:(NSData *)_data onTopic:(NSString *)_topic qos:(MQTTQosLevel)_qos retained:(BOOL)_retained mid:(unsigned int)_mid;
//-----------------------------------------------------------
- (BOOL)Connect :(MarsClient *)_client;
- (BOOL)Subscribe :(NSString *)_topic :(void(^)(NSData *))_handler;
//-----------------------------------------------------------
@end
//-----------------------------------------------------------
#endif
