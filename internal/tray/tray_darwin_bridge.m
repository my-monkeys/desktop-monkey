// Implementation Cocoa du pont declare dans tray_darwin_bridge.h.
//
// Une icone du singe s'installe dans la barre des menus (NSStatusItem). Son
// menu offre : lancer au demarrage (ecrit/retire un LaunchAgent dans
// ~/Library/LaunchAgents), ouvrir les reglages (config.json), et quitter.
//
// Le "quitter" ne fait que lever un drapeau : la boucle principale, ecrite en
// Go, le lit a chaque image et s'arrete proprement (fermeture des fenetres).
#import <Cocoa/Cocoa.h>
#import "tray_darwin_bridge.h"

static int gQuit = 0;

@interface SingeTray : NSObject <NSMenuDelegate>
@property(strong) NSStatusItem *item;
@property(copy) NSString *label;
@property(copy) NSString *exe;
@property(copy) NSString *config;
@end

static SingeTray *gTray = nil;

@implementation SingeTray

- (NSString *)plistPath {
    NSString *dir = [NSHomeDirectory() stringByAppendingPathComponent:@"Library/LaunchAgents"];
    return [dir stringByAppendingPathComponent:
            [self.label stringByAppendingString:@".plist"]];
}

- (BOOL)auDemarrage {
    return [[NSFileManager defaultManager] fileExistsAtPath:[self plistPath]];
}

- (void)basculerDemarrage:(id)sender {
    NSFileManager *fm = [NSFileManager defaultManager];
    NSString *path = [self plistPath];
    if ([self auDemarrage]) {
        [fm removeItemAtPath:path error:nil];
        return;
    }
    NSString *dir = [path stringByDeletingLastPathComponent];
    [fm createDirectoryAtPath:dir withIntermediateDirectories:YES
                   attributes:nil error:nil];
    // launchd charge tout plist de ce dossier a l'ouverture de session
    NSDictionary *plist = @{
        @"Label" : self.label,
        @"ProgramArguments" : @[ self.exe ],
        @"RunAtLoad" : @YES,
    };
    [plist writeToFile:path atomically:YES];
}

- (void)ouvrirReglages:(id)sender {
    [[NSWorkspace sharedWorkspace] openURL:[NSURL fileURLWithPath:self.config]];
}

- (void)quitter:(id)sender {
    gQuit = 1;
}

// met a jour la coche du "lancer au demarrage" chaque fois que le menu s'ouvre
- (void)menuNeedsUpdate:(NSMenu *)menu {
    NSMenuItem *dem = [menu itemAtIndex:0];
    [dem setState:([self auDemarrage] ? NSControlStateValueOn : NSControlStateValueOff)];
}

@end

static NSImage *imageDepuisRGBA(unsigned char *pix, int w, int h) {
    CFDataRef data = CFDataCreate(NULL, pix, (CFIndex)(w * h * 4));
    CGDataProviderRef prov = CGDataProviderCreateWithCFData(data);
    CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
    CGImageRef cg = CGImageCreate(w, h, 8, 32, w * 4, cs,
        kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big,
        prov, NULL, false, kCGRenderingIntentDefault);
    NSImage *img = [[NSImage alloc] initWithCGImage:cg size:NSMakeSize(18, 18)];
    CGImageRelease(cg);
    CGColorSpaceRelease(cs);
    CGDataProviderRelease(prov);
    CFRelease(data);
    return img;
}

void tray_init(const char *nomApp, const char *exe, const char *config,
               unsigned char *icon, int iw, int ih) {
    @autoreleasepool {
        [NSApplication sharedApplication];

        gTray = [SingeTray new];
        gTray.label = [NSString stringWithUTF8String:nomApp];
        gTray.exe = [NSString stringWithUTF8String:exe];
        gTray.config = [NSString stringWithUTF8String:config];

        gTray.item = [[NSStatusBar systemStatusBar]
            statusItemWithLength:NSVariableStatusItemLength];
        if (icon != NULL) {
            gTray.item.button.image = imageDepuisRGBA(icon, iw, ih);
        } else {
            gTray.item.button.title = @"🐒";
        }

        NSMenu *menu = [[NSMenu alloc] init];
        menu.delegate = gTray;
        NSMenuItem *dem = [[NSMenuItem alloc]
            initWithTitle:@"Launch at startup"
                   action:@selector(basculerDemarrage:) keyEquivalent:@""];
        [dem setTarget:gTray];
        [menu addItem:dem];
        NSMenuItem *reg = [[NSMenuItem alloc]
            initWithTitle:@"Open settings..."
                   action:@selector(ouvrirReglages:) keyEquivalent:@""];
        [reg setTarget:gTray];
        [menu addItem:reg];
        [menu addItem:[NSMenuItem separatorItem]];
        NSMenuItem *quit = [[NSMenuItem alloc]
            initWithTitle:@"Quit"
                   action:@selector(quitter:) keyEquivalent:@""];
        [quit setTarget:gTray];
        [menu addItem:quit];

        gTray.item.menu = menu;
    }
}

int tray_quit_requested(void) { return gQuit; }

void tray_fermer(void) {
    @autoreleasepool {
        if (gTray != nil && gTray.item != nil) {
            [[NSStatusBar systemStatusBar] removeStatusItem:gTray.item];
            gTray.item = nil;
        }
    }
}
