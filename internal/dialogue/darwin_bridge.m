// Implementation Cocoa du pont declare dans darwin_bridge.h : une NSWindow
// titree portant une WKWebView plein cadre, chargee sur la page de reglages
// locale. La boucle principale (calque_pump) traite ses evenements comme ceux
// de n'importe quelle fenetre.
#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>
#import "darwin_bridge.h"

static NSWindow *gFenReglages = nil;

@interface DialogueReglages : NSObject <NSWindowDelegate>
@end

static DialogueReglages *gDelegue = nil;

@implementation DialogueReglages
- (void)windowWillClose:(NSNotification *)n {
    gFenReglages = nil;
}
@end

void dialogue_ouvrir(const char *url, const char *titre, int larg, int haut) {
    @autoreleasepool {
        if (gFenReglages != nil) {
            [NSApp activateIgnoringOtherApps:YES];
            [gFenReglages makeKeyAndOrderFront:nil];
            return;
        }
        if (gDelegue == nil) gDelegue = [DialogueReglages new];

        NSRect cadre = NSMakeRect(0, 0, larg, haut);
        gFenReglages = [[NSWindow alloc]
            initWithContentRect:cadre
                      styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable
                        backing:NSBackingStoreBuffered
                          defer:NO];
        gFenReglages.title = [NSString stringWithUTF8String:titre];
        gFenReglages.releasedWhenClosed = NO;
        gFenReglages.delegate = gDelegue;

        WKWebView *web = [[WKWebView alloc] initWithFrame:cadre];
        web.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
        NSURL *u = [NSURL URLWithString:[NSString stringWithUTF8String:url]];
        [web loadRequest:[NSURLRequest requestWithURL:u]];
        gFenReglages.contentView = web;

        [gFenReglages center];
        [NSApp activateIgnoringOtherApps:YES];
        [gFenReglages makeKeyAndOrderFront:nil];
    }
}

void dialogue_fermer(void) {
    @autoreleasepool {
        if (gFenReglages != nil) {
            [gFenReglages close]; // windowWillClose remet gFenReglages a nil
        }
    }
}
