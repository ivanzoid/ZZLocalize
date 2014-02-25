//
//  ZZLocalize.m
//
//  Created by Ivan Zezyulya on 24.02.14.
//  Copyright (c) 2014 ivanzez. All rights reserved.
//

#import "ZZLocalize.h"
#import "CHCSVParser.h"

static NSMutableDictionary *GLocalizationDict;
static NSMutableDictionary *GLocalizationFallbackDict;
static ZZLocalizeOptions GLocalizeOptions;
static NSInteger GLocalizationLanguagesCount;
static NSString * const kDefaultFileName = @"Localization.csv";
static NSString * const kLanguageKey = @"language";

void ZZLocalizeInit(void)
{
    ZZLocalizeInitWithOptions(ZZLocalizeOptionUseFallbackLanguage);
}

extern void ZZLocalizeInitWithOptions(ZZLocalizeOptions options)
{
    ZZLocalizeInitWithOptionsAndFileName(options, kDefaultFileName);
}

static NSString *TrimQuotesFromValue(NSString *value)
{
    return [value stringByTrimmingCharactersInSet:[NSCharacterSet characterSetWithCharactersInString:@"\""]];
}

extern void ZZLocalizeInitWithOptionsAndFileName(ZZLocalizeOptions options, NSString *fileName)
{
    GLocalizationDict = [NSMutableDictionary new];
    GLocalizeOptions = options;

    NSString *filePath = [[NSBundle mainBundle] pathForResource:fileName ofType:nil];
    NSArray *translations = [NSArray arrayWithContentsOfCSVFile:filePath options:CHCSVParserOptionsRecognizesComments];

    if ([translations count] == 0) {
        return;
    }

    NSString *currentLanguage = [NSLocale preferredLanguages][0];

    NSArray *languages = translations[0];
    if ([languages count] == 0) {
        LogCError(@"[ZZLocalize] Bad line with languages.");
        return;
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
        return;
    }

    BOOL useFallbackLanguage = (options & ZZLocalizeOptionUseFallbackLanguage);
    if (useFallbackLanguage) {
        GLocalizationFallbackDict = [NSMutableDictionary new];
    }

    for (NSInteger i = 1; i < [translations count]; i++) {
        NSArray *translation = translations[i];
        if (languageIndex >= [translation count]) {
            LogCWarning(@"[ZZLocalize] Bad line at %@:%d", fileName, i);
            continue;
        }
        NSString *key = translation[0];
        NSString *value = TrimQuotesFromValue(translation[languageIndex]);
        if ([value length]) {
            GLocalizationDict[key] = value;
        }

        if (useFallbackLanguage) {
            NSString *fallbackValue = TrimQuotesFromValue(translation[1]);
            if ([fallbackValue length]) {
                GLocalizationFallbackDict[key] = fallbackValue;
            }
        }
    }
}

NSString * ZZLocalize(NSString *key)
{
    NSString *result = GLocalizationDict[key];

    if (result == nil) {
        result = GLocalizationFallbackDict[key];
        if (result) {
            LogCWarning(@"[ZZLocalize] Missing translation for key '%@'. Using value for default language.", key);
        } else {
            LogCWarning(@"[ZZLocalize] Missing translation for key '%@'.", key);
        }
    }

    if (!result) {
        result = key;
    }

    return result;
}
