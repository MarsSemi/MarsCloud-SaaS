//-----------------------------------------------------------
#import <Cocoa/Cocoa.h>

//-----------------------------------------------------------
#import "MarsClient.h"
#import "MarsMQTT.h"
//-----------------------------------------------------------
//
//-----------------------------------------------------------
MarsClient *_client = nil;
MarsMQTT *_mqtt = nil;
//-----------------------------------------------------------
void testClient(MarsClient *_client)
{
    NSMutableDictionary *_payload = [[NSMutableDictionary alloc] init];
    NSString *_resp = nil;

    _payload[@"temp"] = @(23.7);
    _payload[@"humi"] = @(85);

    if([_client PutMessage :@"test/my/topic" :_payload :nil]) NSLog(@"PutMessage SUCCESS");
    if([_client PutData :@"dev" :@"test" :_payload :nil]) NSLog(@"PutData SUCCESS");

    _resp = [_client GetLastData :@"dev" :@"test" :1 :nil];
    
    if(_resp != nil)
    {
        NSLog(@"GetLastData SUCCESS");
        NSLog(@"%@", _resp);

        NSData *_data = [_resp dataUsingEncoding:NSUTF8StringEncoding];
        NSDictionary *_respJSON = [NSJSONSerialization JSONObjectWithData:_data options:0 error:nil];
        NSArray *_values = (NSArray *)_respJSON[@"results"];
        NSDictionary *_item = _values[0];

        if([_client RemoveData :@"dev" :@"test" :_item[@"ukey"] :nil])
            NSLog(@"RemoveData SUCCESS");
    }

    _resp = [_client CallService :@"service.myService" :@"/api/hello" :nil :nil];
    if(_resp != nil)
        NSLog(@"CallService : %@", _resp);
}
//-----------------------------------------------------------
bool testMQTT(MarsClient *_client)
{
    if([_mqtt Connect :_client])
    {
        NSLog(@"MQTT SUCCESS");
        if([_mqtt Subscribe:@"test/+/#" :nil])
            return true;
    }
    
    return false;
}
//-----------------------------------------------------------
int main(int argc, const char * argv[])
{
    @autoreleasepool
    {
        [MQTTLog setLogLevel :DDLogLevelError];
        
        _client = [MarsClient new];
        _mqtt = [MarsMQTT new];
        
        if([_client Login :@"https://test.mars-cloud.com":@"test":@"test":@"justtest"])
        {
            NSLog(@"Login SUCCESS");
            
            if(testMQTT(_client))
                testClient(_client);
        }
    }
    
    return NSApplicationMain(argc, argv);
}
//-----------------------------------------------------------
