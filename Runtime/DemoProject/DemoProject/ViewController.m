//
//  ViewController.m
//  DemoProject
//
//  Created by Ivan on 24.02.14.
//  Copyright (c) 2014 aldigit. All rights reserved.
//

#import "ViewController.h"
#import "ZZLocalize.h"

@interface ViewController ()
@property (weak, nonatomic) IBOutlet UILabel *helloLabel;
@end

@implementation ViewController

- (void) viewDidLoad
{
    [super viewDidLoad];

    self.helloLabel.text = Localize(@"helloWorld");
}

@end
