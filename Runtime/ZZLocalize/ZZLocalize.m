//
//  ZZLocalize.m
//
//  Created by Ivan Zezyulya on 24.02.14.
//  Copyright (c) 2014 ivanzez. All rights reserved.
//

#import "ZZLocalize.h"
#import "CHCSVParser.h"
#import "ZZLocalizePrivate.h"

static NSMutableDictionary *GLocalizationDict;
static NSMutableDictionary *GLocalizationFallbackDict;
static NSInteger GLocalizationLanguagesCount;
static NSString * const kDefaultFileName = @"Localization.csv";
static NSString * const kLanguageKey = @"language";
static NSString * const kNSErrorDomainName = @"ZZLocalize";

BOOL ZZLocalizeInit(NSError **error)
{
    return ZZLocalizeInitWithFileName(kDefaultFileName, error);
}

static NSString *TrimQuotesFromValue(NSString *value)
{
    NSString *result = value;
    result = [result stringByReplacingOccurrencesOfString:@"\"\"" withString:@"\""];
    result = [result stringByTrimmingCharactersInSet:[NSCharacterSet characterSetWithCharactersInString:@"\""]];
    return result;
}

static NSError *ErrorWithDescription(NSString *description)
{
    NSCParameterAssert(description);

    NSError *error = [NSError errorWithDomain:kNSErrorDomainName code:0 userInfo:@{NSLocalizedDescriptionKey: description}];
    return error;
}

BOOL ZZLocalizeInitWithFileName(NSString *fileName, NSError **error)
{
    GLocalizationDict = [NSMutableDictionary new];

    NSString *filePath = [[NSBundle mainBundle] pathForResource:fileName ofType:nil];
    if (filePath == nil) {
        if (error) {
            *error = ErrorWithDescription([NSString stringWithFormat:@"Can't find %@.", fileName]);
        }
        return NO;
    }

    NSArray *translations = [NSArray arrayWithContentsOfCSVFile:filePath options:CHCSVParserOptionsRecognizesComments];

    if ([translations count] == 0) {
        return YES;
    }

    NSString *currentLanguage = [NSLocale preferredLanguages][0];

    NSArray *languages = translations[0];
    if ([languages count] == 0) {
        if (error) {
            *error = ErrorWithDescription(@"Bad line with languages.");
        }
        return NO;
    }

    GLocalizationLanguagesCount = [languages count] - 1;

    NSInteger languageIndex = -1;
    for (NSInteger i = 1; i < [languages count]; i++) {
        NSString *languageValue = languages[i];
        if ([currentLanguage hasPrefix:languageValue]) {
            languageIndex = i;
            break;
        }
    }

    if (languageIndex == -1) {
        return YES;
    }

    GLocalizationFallbackDict = [NSMutableDictionary new];

    for (NSInteger i = 1; i < [translations count]; i++) {
        NSArray *translation = translations[i];
        if (languageIndex >= [translation count]) {
            ZZLocalizeCWarn("bad line at %@:%d", fileName, i);
            continue;
        }
        NSString *key = translation[0];
        NSString *value = TrimQuotesFromValue(translation[languageIndex]);
        if ([value length]) {
            GLocalizationDict[key] = value;
        }

        NSString *fallbackValue = TrimQuotesFromValue(translation[1]);
        if ([fallbackValue length]) {
            GLocalizationFallbackDict[key] = fallbackValue;
        }
    }

    return YES;
}

NSString * ZZLocalize(NSString *key)
{
    NSString *result = GLocalizationDict[key];

    if (result == nil) {
        result = GLocalizationFallbackDict[key];
        if (result) {
            ZZLocalizeCWarn(@"missing translation for key '%@'. Using value for default language.", key);
        } else {
            ZZLocalizeCWarn(@"missing translation for key '%@'.", key);
        }
    }

    if (!result) {
        result = key;
    }

    return result;
}
