//
//  ZZLocalizePrivate.h
//  iHerb
//
//  Created by Ivan Zezyulya on 09.04.14.
//  Copyright (c) 2014 aldigit. All rights reserved.
//

#define ZZLocalizeLogPrefix @"[ZZLocalize] "

#define ZZLocalizeVerbose(fmt, ...) DDLogVerbose(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeDebug(fmt, ...)   DDLogDebug(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeInfo(fmt, ...)    DDLogInfo(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeWarn(fmt, ...)    DDLogWarn(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeError(fmt, ...)   DDLogError(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)

#define ZZLocalizeCVerbose(fmt, ...) DDLogCVerbose(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeCDebug(fmt, ...)   DDLogCDebug(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeCInfo(fmt, ...)    DDLogCInfo(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeCWarn(fmt, ...)    DDLogCWarn(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
#define ZZLocalizeCError(fmt, ...)   DDLogCError(ZZLocalizeLogPrefix fmt, ##__VA_ARGS__)
