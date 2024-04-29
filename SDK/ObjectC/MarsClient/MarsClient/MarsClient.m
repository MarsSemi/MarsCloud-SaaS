//-----------------------------------------------------------
#import "MarsClient.h"
//-----------------------------------------------------------
@implementation MarsClient
//-----------------------------------------------------------
- (BOOL)HttpRequest :(NSString *)_api :(NSData *)_payload :(void(^)(NSData *))_handler
{
    @try
    {
        NSString *_urlText = [_Host stringByAppendingString:_api];
        NSURL *_url = [NSURL URLWithString: _urlText];
        NSURLSession *_session = [NSURLSession sharedSession];
        NSMutableURLRequest *_req = [NSMutableURLRequest requestWithURL:_url cachePolicy:NSURLRequestReloadIgnoringLocalCacheData timeoutInterval:15.0];

        if(_Token != nil) [_req setValue :[@"Bearer " stringByAppendingString:_Token] forHTTPHeaderField:@"Authentication"];

        [_req setValue :@"application/json" forHTTPHeaderField:@"Content-Type"];
        [_req setHTTPMethod :_payload == nil ? @"GET" : @"POST"];
        [_req setHTTPBody :_payload];

        [[_session dataTaskWithRequest:_req completionHandler :^(NSData *_data, NSURLResponse *_resp, NSError *_error)
        {
            @try
            {
                NSHTTPURLResponse *_httpResp = (NSHTTPURLResponse *)_resp;
                //NSLog(@"???? : %@", [[NSString alloc] initWithData :_data encoding:NSUTF8StringEncoding]);
                if(_handler != nil)
                    _handler(_httpResp.statusCode == 200 && _data.length > 0 ? _data : nil);
            }
            @catch (NSException *_e) { _handler(nil); }

        }] resume];

        return YES;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return NO;
}
//-----------------------------------------------------------
- (NSString *)GetUser
{
    return _User;
}
//-----------------------------------------------------------
- (NSString *)GetToken
{
    return _Token;
}
//-----------------------------------------------------------
- (BOOL)Login :(NSString *)_host :(NSString *)_user :(NSString *)_pwd :(NSString *)_proj
{
    @try
    {
        _Host = _host;
        _User = _user;
        _Password = _pwd;
        _Proj = _proj;  
        _Token = nil;

        NSMutableDictionary *_payload = [[NSMutableDictionary alloc]init];
        NSData *_respone = nil;

        _payload[@"usr"] = _User;
        _payload[@"pwd"] = _Password;
        _payload[@"proj"] = _Proj;

        dispatch_semaphore_t _sync = dispatch_semaphore_create(0);

        [self HttpRequest :@"/auth/login?" :[NSJSONSerialization dataWithJSONObject:_payload options:NSJSONWritingPrettyPrinted error:nil] :^(NSData *_data)
        {
            @try
            {
                 _Token = [[NSString alloc] initWithData :_data encoding:NSUTF8StringEncoding];
            }
            @catch (NSException *_e) {}
            @finally { dispatch_semaphore_signal(_sync); }
        }];
        
        dispatch_semaphore_wait(_sync, dispatch_time(DISPATCH_TIME_NOW, 15*NSEC_PER_SEC));
        return _Token != nil;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return NO;
}
//-----------------------------------------------------------
- (BOOL)PutData :(NSString *)_uuid :(NSString *)_suid :(NSDictionary *)_payload :(void(^)(BOOL))_handler
{
    @try
    {
        dispatch_semaphore_t _sync = nil;
        BOOL __block _result = YES;

        NSMutableDictionary *_data = [[NSMutableDictionary alloc]init];
        NSMutableArray *_values = [[NSMutableArray alloc] init];

        [_values addObject :_payload];

        _data[@"uuid"] = _uuid;
        _data[@"suid"] = _suid;
        _data[@"values"] = _values;

        if(_handler == nil) { _sync = dispatch_semaphore_create(0); _result = NO; }

        [self HttpRequest :@"/api/put?data" :[NSJSONSerialization dataWithJSONObject:_data options:NSJSONWritingPrettyPrinted error:nil] :^(NSData *_data)
        {
            @try
            {
                _result = (_data != nil);
                if(_handler != nil)
                    _handler(_result);
                
            }
            @catch (NSException *_e) {}
            @finally { dispatch_semaphore_signal(_sync); }
        }];
        
        if(_sync != nil) dispatch_semaphore_wait(_sync, dispatch_time(DISPATCH_TIME_NOW, 15*NSEC_PER_SEC));

        return _result;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return NO;
}
//-----------------------------------------------------------
- (NSString *)GetLastData :(NSString *)_uuid :(NSString *)_suid :(int)_count :(void(^)(NSString *))_handler
{
    @try
    {
        dispatch_semaphore_t _sync = nil;
        NSString __block *_result = nil;
        NSMutableDictionary *_data = [[NSMutableDictionary alloc]init];

        _data[@"uuid"] = _uuid;
        _data[@"suid"] = _suid;
        _data[@"count"] = @(_count);

        if(_handler == nil) { _sync = dispatch_semaphore_create(0); _result = nil; }

        [self HttpRequest :@"/api/lastdata?method=read" :[NSJSONSerialization dataWithJSONObject:_data options:NSJSONWritingPrettyPrinted error:nil] :^(NSData *_data)
        {
            @try
            {
                _result =  _data != nil ? [[NSString alloc] initWithData :_data encoding:NSUTF8StringEncoding] : nil;
                if(_handler != nil)
                    _handler(_result);
            }
            @catch (NSException *_e) {}
            @finally { dispatch_semaphore_signal(_sync); }
        }];
        
        if(_sync != nil) dispatch_semaphore_wait(_sync, dispatch_time(DISPATCH_TIME_NOW, 15*NSEC_PER_SEC));

        return _result;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return nil;
}
//-----------------------------------------------------------
- (BOOL)RemoveData :(NSString *)_uuid :(NSString *)_suid :(NSString *)_ukey :(void(^)(BOOL))_handler
{
    @try
    {
        dispatch_semaphore_t _sync = nil;
        BOOL __block _result = YES;

        NSMutableDictionary *_data = [[NSMutableDictionary alloc]init];

        _data[@"uuid"] = _uuid;
        _data[@"suid"] = _suid;
        _data[@"ukey"] = _ukey;

        if(_handler == nil) { _sync = dispatch_semaphore_create(0); _result = NO; }

        [self HttpRequest :@"/api/del?data" :[NSJSONSerialization dataWithJSONObject:_data options:NSJSONWritingPrettyPrinted error:nil] :^(NSData *_data)
        {
            @try
            {
                _result = (_data != nil);
                if(_handler != nil)
                    _handler(_result);                
            }
            @catch (NSException *_e) {}
            @finally { dispatch_semaphore_signal(_sync); }
        }];
        
        if(_sync != nil) dispatch_semaphore_wait(_sync, dispatch_time(DISPATCH_TIME_NOW, 15*NSEC_PER_SEC));

        return _result;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return NO;
}
//-----------------------------------------------------------
- (BOOL)PutMessage :(NSString *)_topic :(NSDictionary *)_payload :(void(^)(BOOL))_handler
{
    @try
    {
        dispatch_semaphore_t _sync = nil;
        BOOL __block _result = YES;

        if(_handler == nil) { _sync = dispatch_semaphore_create(0); _result = NO; }
        
        _topic = [@"/api/put?message&topic=" stringByAppendingString:_topic];

        [self HttpRequest :_topic :[NSJSONSerialization dataWithJSONObject:_payload options:NSJSONWritingPrettyPrinted error:nil] :^(NSData *_data)
        {
            @try
            {
                _result = (_data != nil);
                if(_handler != nil)
                    _handler(_result);
            }
            @catch (NSException *_e) {}
            @finally { dispatch_semaphore_signal(_sync); }
        }];
        
        if(_sync != nil) dispatch_semaphore_wait(_sync, dispatch_time(DISPATCH_TIME_NOW, 15*NSEC_PER_SEC));

        return _result;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return NO;
}
//-----------------------------------------------------------
- (NSString *)CallService :(NSString *)_service :(NSString *)_api :(NSString *)_payload :(void(^)(NSString *))_handler
{
    @try
    {
        dispatch_semaphore_t _sync = nil;
        NSString __block *_result = nil;

        _api = [_service stringByAppendingString:_api];
        _api = [@"/services/" stringByAppendingString:_api];

        if(_handler == nil) { _sync = dispatch_semaphore_create(0); _result = nil; }

        [self HttpRequest :_api :[_payload dataUsingEncoding:NSUTF8StringEncoding] :^(NSData *_data)
        {
            @try
            {
                _result =  _data != nil ? [[NSString alloc] initWithData :_data encoding:NSUTF8StringEncoding] : nil;
                if(_handler != nil)
                    _handler(_result);
            }
            @catch (NSException *_e) {}
            @finally { dispatch_semaphore_signal(_sync); }
        }];
        
        if(_sync != nil) dispatch_semaphore_wait(_sync, dispatch_time(DISPATCH_TIME_NOW, 15*NSEC_PER_SEC));

        return _result;
    }
    @catch (NSException *_e) { NSLog(@"Error : %@", _e.reason); }
    @finally {}
    return nil;
}
//-----------------------------------------------------------
@end
//-----------------------------------------------------------
