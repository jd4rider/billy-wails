// tray_darwin.m — compiled only on macOS (CGO picks up _darwin suffix).
// Creates an NSStatusItem via dispatch_async so it doesn't conflict with
// Wails' NSApp main run loop (avoids calling dispatch_main()).

#import <Cocoa/Cocoa.h>
#include "_cgo_export.h"

@interface _BillyTrayDelegate : NSObject
- (void)doShow:(id)sender;
- (void)doQuit:(id)sender;
@end

@implementation _BillyTrayDelegate
- (void)doShow:(id)sender { trayShowCallback(); }
- (void)doQuit:(id)sender { trayQuitCallback(); }
@end

static NSStatusItem       *_billyItem;
static _BillyTrayDelegate *_billyDelegate;

void billyInitTray(const unsigned char *iconBytes, int iconLen) {
    // Build NSData synchronously — copies bytes before dispatch_async so the
    // caller can free the buffer immediately after this function returns.
    NSData *imgData = [NSData dataWithBytes:iconBytes length:(NSUInteger)iconLen];
    dispatch_async(dispatch_get_main_queue(), ^{
        NSImage *img = [[NSImage alloc] initWithData:imgData];
        img.template = YES;  // auto-inverts for dark/light menu bar

        _billyDelegate = [_BillyTrayDelegate new];

        _billyItem = [[NSStatusBar systemStatusBar]
                        statusItemWithLength:NSSquareStatusItemLength];
        _billyItem.button.image = img;
        _billyItem.button.toolTip = @"Billy — local AI assistant";

        NSMenu *menu = [NSMenu new];

        NSMenuItem *show = [[NSMenuItem alloc]
            initWithTitle:@"Show Billy"
            action:@selector(doShow:)
            keyEquivalent:@""];
        show.target = _billyDelegate;
        [menu addItem:show];

        [menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quit = [[NSMenuItem alloc]
            initWithTitle:@"Quit Billy"
            action:@selector(doQuit:)
            keyEquivalent:@""];
        quit.target = _billyDelegate;
        [menu addItem:quit];

        _billyItem.menu = menu;
    });
}
