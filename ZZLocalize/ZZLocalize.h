//
//  ZZLocalize.h
//
//  Created by Ivan Zezyulya on 24.02.14.
//  Copyright (c) 2014 ivanzez. All rights reserved.
//

#import <Foundation/Foundation.h>

extern void ZZLocalizeInit(void);
extern void ZZLocalizeInitWithFileName(NSString *fileName);
extern NSString * ZZLocalize(NSString *key);

#define Localize(string) ZZLocalize(string)
