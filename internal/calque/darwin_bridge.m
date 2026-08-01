// Implementation Cocoa du pont declare dans darwin_bridge.h.
//
// La fenetre est l'equivalent macOS de la fenetre en couche Windows : sans
// bordure, a fond transparent, toujours au premier plan et traversante. Son
// contenu est une image RGBA posee sur le calque (CALayer) d'une vue.
//
// Traversante : NSWindow ne teste pas les clics au pixel pres comme Windows.
// On met donc ignoresMouseEvents = YES (tout traverse). Le singe reagit quand
// meme, car l'application lit la souris globalement (voir internal/souris) ;
// seul effet de bord : cliquer le singe atteint aussi ce qu'il y a derriere.
#import <Cocoa/Cocoa.h>
#import "darwin_bridge.h"

void calque_init(void) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        // Accessory : pas d'icone dans le Dock, ne vole pas le focus.
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
        [NSApp finishLaunching];
    }
}

void* calque_ouvrir(int w, int h) {
    @autoreleasepool {
        NSRect frame = NSMakeRect(0, 0, w, h);
        NSWindow* win = [[NSWindow alloc]
            initWithContentRect:frame
                      styleMask:NSWindowStyleMaskBorderless
                        backing:NSBackingStoreBuffered
                          defer:NO];
        [win setOpaque:NO];
        [win setBackgroundColor:[NSColor clearColor]];
        [win setHasShadow:NO];
        [win setLevel:NSStatusWindowLevel];
        [win setIgnoresMouseEvents:YES];
        // visible sur tous les bureaux, immobile lors des changements d'espace,
        // absente du cycle Cmd+Tab
        [win setCollectionBehavior:
            NSWindowCollectionBehaviorCanJoinAllSpaces |
            NSWindowCollectionBehaviorStationary |
            NSWindowCollectionBehaviorIgnoresCycle];

        NSView* v = [[NSView alloc] initWithFrame:frame];
        [v setWantsLayer:YES];
        v.layer.magnificationFilter = @"nearest"; // kCAFilterNearest : pixel art net
        [win setContentView:v];
        [win orderFrontRegardless];
        return (void*)CFBridgingRetain(win);
    }
}

void calque_afficher(void* winp, unsigned char* pix, int w, int h,
                     int x, int y, int screenH) {
    @autoreleasepool {
        NSWindow* win = (__bridge NSWindow*)winp;

        // Copie des pixels dans une CFData que possede l'image : le calque peut
        // composer plus tard, il ne doit pas dependre du tampon Go reutilise.
        CFDataRef data = CFDataCreate(NULL, pix, (CFIndex)(w * h * 4));
        CGDataProviderRef prov = CGDataProviderCreateWithCFData(data);
        CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
        CGImageRef img = CGImageCreate(
            w, h, 8, 32, w * 4, cs,
            kCGImageAlphaPremultipliedLast | kCGBitmapByteOrder32Big,
            prov, NULL, false, kCGRenderingIntentDefault);

        [win contentView].layer.contents = (__bridge id)img;

        CGImageRelease(img);
        CGColorSpaceRelease(cs);
        CGDataProviderRelease(prov);
        CFRelease(data);

        // x, y sont dans l'espace unifie de tous les ecrans (origine en haut a
        // gauche de l'union). On repasse en repere Cocoa : origine en bas a
        // gauche de l'ecran principal, y vers le haut. Ainsi le singe s'affiche
        // correctement sur n'importe quel ecran, pas seulement le principal.
        int ox = 0, oy = 0;
        uint32_t n = 0;
        CGGetActiveDisplayList(0, NULL, &n);
        if (n > 16) n = 16;
        if (n > 0) {
            CGDirectDisplayID ids[16];
            CGGetActiveDisplayList(n, ids, &n);
            CGRect u = CGDisplayBounds(ids[0]);
            for (uint32_t i = 1; i < n; i++) u = CGRectUnion(u, CGDisplayBounds(ids[i]));
            ox = (int)u.origin.x;
            oy = (int)u.origin.y;
        }
        int mainH = (int)CGDisplayBounds(CGMainDisplayID()).size.height;
        (void)screenH;
        [win setFrameOrigin:NSMakePoint(ox + x, mainH - (oy + y) - h)];
    }
}

void calque_traversant(void* winp, int oui) {
    @autoreleasepool {
        NSWindow* win = (__bridge NSWindow*)winp;
        [win setIgnoresMouseEvents:(oui ? YES : NO)];
    }
}

void calque_devant(void* winp) {
    @autoreleasepool {
        NSWindow* win = (__bridge NSWindow*)winp;
        [win orderFrontRegardless]; // au-dessus des crottes de meme niveau
    }
}

int calque_pump(void) {
    @autoreleasepool {
        NSEvent* e;
        while ((e = [NSApp nextEventMatchingMask:NSEventMaskAny
                                       untilDate:[NSDate distantPast]
                                          inMode:NSDefaultRunLoopMode
                                         dequeue:YES]) != nil) {
            [NSApp sendEvent:e];
        }
    }
    return 1;
}

void calque_fermer(void* winp) {
    @autoreleasepool {
        NSWindow* win = (NSWindow*)CFBridgingRelease(winp);
        [win close];
    }
}

void calque_ecran(int* w, int* h) {
    // taille de l'union de tous les ecrans (le monde du singe)
    uint32_t n = 0;
    CGGetActiveDisplayList(0, NULL, &n);
    if (n == 0) { *w = *h = 0; return; }
    if (n > 16) n = 16;
    CGDirectDisplayID ids[16];
    CGGetActiveDisplayList(n, ids, &n);
    CGRect u = CGDisplayBounds(ids[0]);
    for (uint32_t i = 1; i < n; i++) u = CGRectUnion(u, CGDisplayBounds(ids[i]));
    *w = (int)u.size.width;
    *h = (int)u.size.height;
}
