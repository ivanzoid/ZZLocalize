//
//  ZZLocalize.h
//
//  Created by Ivan Zezyulya on 24.02.14.
//  Copyright (c) 2014 ivanzez. All rights reserved.
//

#import <Foundation/Foundation.h>

/// @note Same as ZZLocalizeInitWithFileName with fileName = @"Localization.csv".
extern BOOL ZZLocalizeInit(NSError **error);
extern BOOL ZZLocalizeInitWithFileName(NSString *fileName, NSError **error);

extern NSString * ZZLocalize(NSString *key);

#ifndef ZZLOCALIZE_NO_SHORTHAND
#define Localize(string) ZZLocalize(string)
#endif
