//-----------------------------------------------------------
#import "MarsMQTT.h"
//-----------------------------------------------------------
@implementation MarsMQTT
//-----------------------------------------------------------
//
//-----------------------------------------------------------
- (void)connected :(MQTTSession *)_session { _MQTTClient = _session; }
//-----------------------------------------------------------
- (void)connectionClosed :(MQTTSession *)_session { _MQTTClient = nil; }
//-----------------------------------------------------------
- (void)newMessage:(MQTTSession *)_session data:(NSData *)_data onTopic:(NSString *)_topic qos:(MQTTQosLevel)_qos retained:(BOOL)_retained mid:(unsigned int)_mid
{
    @try
    {
        NSLog(@"Topic : %@", _topic);
        NSLog(@"%@", [[NSString alloc] initWithData :_data encoding:NSUTF8StringEncoding]);
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
}
//-----------------------------------------------------------
//
//-----------------------------------------------------------
- (BOOL)Connect :(MarsClient *)_client;
{
    @try
    {
        int _tick = 0;
        MQTTSession *_session = [[MQTTSession alloc] init];
        NSString *_id = @"";
        
        _MQTTClient = nil;
        _MarsClient = _client;
        
        _id = [_id stringByAppendingString :[_MarsClient GetToken]];
        _id = [_id stringByAppendingString :[NSString stringWithFormat:@"@%d", rand()%1000]];
        
        //MQTTWebsocketTransport *_train = [[MQTTWebsocketTransport alloc] init];
        //_train.url = @"wss://test.mars-cloud.com:8884";
        
        MQTTCFSocketTransport *_train = [[MQTTCFSocketTransport alloc] init];
        _train.host = @"test.mars-cloud.com";
        _train.port = 8883;
        _train.tls = YES;
        
        _session.transport = _train;
        _session.delegate = self;

        [_session setClientId :_id];
        [_session setUserName :[_MarsClient GetUser]];
        [_session setPassword :[_MarsClient GetToken]];
        [_session connectAndWaitTimeout :10];
        
        return _MQTTClient != nil;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return NO;
}
//-----------------------------------------------------------
- (BOOL)Subscribe :(NSString *)_topic :(void(^)(NSData *))_handler
{
    @try
    {
        if(_MQTTClient != nil)
        {
            [_MQTTClient subscribeToTopic:_topic atLevel:0 subscribeHandler:^(NSError *_error, NSArray<NSNumber *> *_gQoss)
            {
                if(_error)
                    NSLog(@"Subscribe Error : %@", _error);
                else
                    NSLog(@"Subscribe Success : %@", _topic);
            }];
            
            return YES;
        }
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return NO;
}
//-----------------------------------------------------------
@end
//-----------------------------------------------------------
