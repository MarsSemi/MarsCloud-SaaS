#ifndef _MARSCLIENT_
#define _MARSCLIENT_
//-----------------------------------------------------------
#pragma once
//-----------------------------------------------------------
#import <Foundation/Foundation.h>
//-----------------------------------------------------------
@interface MarsClient : NSObject
{
    NSString *_Host;
    NSString *_User;
    NSString *_Password;
    NSString *_Proj;
    NSString *_Token;
}
//-----------------------------------------------------------
- (BOOL)HttpRequest :(NSString *)_api :(NSData *)_payload :(void(^)(NSData *))_handler;
- (BOOL)Login :(NSString *)_host :(NSString *)_user :(NSString *)_pwd :(NSString *)_proj;

- (BOOL)PutData :(NSString *)_uuid :(NSString *)_suid :(NSDictionary *)_payload :(void(^)(BOOL))_handler;
- (NSString *)GetLastData :(NSString *)_uuid :(NSString *)_suid :(int)_count :(void(^)(NSString *))_handler;
- (BOOL)RemoveData :(NSString *)_uuid :(NSString *)_suid :(NSString *)_ukey :(void(^)(BOOL))_handler;

- (NSString *)CallService :(NSString *)_service :(NSString *)_api :(NSString *)_payload :(void(^)(NSString *))_handler;
//-----------------------------------------------------------
@end
//-----------------------------------------------------------
#endif