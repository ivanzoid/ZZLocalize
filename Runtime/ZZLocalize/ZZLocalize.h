//
//  ZZLocalize.h
//
//  Created by Ivan Zezyulya on 24.02.14.
//  Copyright (c) 2014 ivanzez. All rights reserved.
//

#import <Foundation/Foundation.h>

typedef NS_ENUM(NSInteger, ZZLocalizeOptions) {
    /// If set, "default" language will be loaded and used as fallback language when no value for current language is present in translation.
    /// "Default" language is first language listed in translation file.
    ZZLocalizeOptionUseFallbackLanguage = (0 << 0)
};

/// @note Same as ZZLocalizeInitWithOptions with options = ZZLocalizeOptionUseFallbackLanguage.
extern void ZZLocalizeInit(void);

/// @note Same as ZZLocalizeInitWithOptionsAndFileName with fileName = @"Localization.csv".
extern void ZZLocalizeInitWithOptions(ZZLocalizeOptions options);

extern void ZZLocalizeInitWithOptionsAndFileName(ZZLocalizeOptions options, NSString *fileName);

extern NSString * ZZLocalize(NSString *key);

#define Localize(string) ZZLocalize(string)
