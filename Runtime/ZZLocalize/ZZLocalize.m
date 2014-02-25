//
//  ZZLocalize.m
//
//  Created by Ivan Zezyulya on 24.02.14.
//  Copyright (c) 2014 ivanzez. All rights reserved.
//

#import "ZZLocalize.h"
#import "CHCSVParser.h"

static NSMutableDictionary *ZZLocalizationDict;
static NSString * const kDefaultFileName = @"Localization.csv";
static NSString * const kLanguageKey = @"language";

void ZZLocalizeInit(void)
{
    ZZLocalizeInitWithFileName(kDefaultFileName);
}

void ZZLocalizeInitWithFileName(NSString *fileName)
{
    ZZLocalizationDict = [NSMutableDictionary new];

    NSString *filePath = [[NSBundle mainBundle] pathForResource:fileName ofType:nil];
    NSArray *translations = [NSArray arrayWithContentsOfCSVFile:filePath options:CHCSVParserOptionsSanitizesFields | CHCSVParserOptionsRecognizesComments];

    if ([translations count] == 0) {
        return;
    }

    NSString *currentLocaleISOCode = [NSLocale preferredLanguages][0];

    NSArray *languages = translations[0];
    if ([languages count] == 0) {
        NSLog(@"[ZZLocalize] Bad line with languages.");
        return;
    }

    NSInteger languageIndex = -1;
    for (NSInteger i = 1; i < [languages count]; i++) {
        NSString *languageValue = languages[i];
        if ([languageValue hasPrefix:currentLocaleISOCode]) {
            languageIndex = i;
            break;
        }
    }

    if (languageIndex == -1) {
        return;
    }

    for (NSInteger i = 1; i < [translations count]; i++) {
        NSArray *translation = translations[i];
        if (languageIndex >= [translation count]) {
            NSLog(@"[ZZLocalize] Bad line at %@:%d", fileName, i);
            continue;
        }
        NSString *key = translation[0];
        NSString *value = translation[languageIndex];
        ZZLocalizationDict[key] = value;
    }
}

NSString * ZZLocalize(NSString *key)
{
    NSString *result = ZZLocalizationDict[key];
    return result;
}
