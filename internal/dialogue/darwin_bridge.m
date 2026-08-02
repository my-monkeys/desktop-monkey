// Implementation Cocoa du pont declare dans darwin_bridge.h.
//
// La fenetre de reglages est une vraie boite de dialogue native : NSTabView
// pour les onglets, NSSlider pour les curseurs, NSPopUpButton pour les choix,
// NSButton pour les cases. Elle est construite a partir d'une description JSON
// pour que le Go n'ait rien a connaitre d'AppKit. "Enregistrer" serialise les
// valeurs en JSON, que le Go recolte via dialogue_resultat().
#import <Cocoa/Cocoa.h>
#import "darwin_bridge.h"
#include <string.h>

static NSWindow *gFen = nil;
static char *gResultat = NULL;
static NSMutableDictionary *gCtrl = nil;  // cle -> controle
static NSMutableDictionary *gType = nil;  // cle -> type de champ
static NSMutableDictionary *gVal = nil;   // cle -> label de valeur (curseurs)
static NSMutableDictionary *gPas = nil;   // cle -> pas (format d'affichage)
static NSMutableDictionary *gOpts = nil;  // cle -> options d'un choix

@interface DialogueSinge : NSObject <NSWindowDelegate>
@end

static DialogueSinge *gCible = nil;

static NSString *formatValeur(double v, double pas) {
    if (pas >= 1) return [NSString stringWithFormat:@"%.0f", v];
    return [NSString stringWithFormat:@"%.2f", v];
}

@implementation DialogueSinge

- (void)curseurChange:(NSSlider *)s {
    NSString *cle = s.identifier;
    NSTextField *l = gVal[cle];
    if (l != nil) {
        l.stringValue = formatValeur(s.doubleValue, [gPas[cle] doubleValue]);
    }
}

- (void)enregistrer:(id)sender {
    NSMutableDictionary *d = [NSMutableDictionary new];
    for (NSString *cle in gCtrl) {
        NSString *type = gType[cle];
        id c = gCtrl[cle];
        if ([type isEqualToString:@"curseur"]) {
            d[cle] = @([(NSSlider *)c doubleValue]);
        } else if ([type isEqualToString:@"entier"]) {
            d[cle] = @((long)llround([(NSSlider *)c doubleValue]));
        } else if ([type isEqualToString:@"case"]) {
            d[cle] = @([(NSButton *)c state] == NSControlStateValueOn);
        } else if ([type isEqualToString:@"choix"]) {
            NSArray *opts = gOpts[cle];
            NSInteger i = [(NSPopUpButton *)c indexOfSelectedItem];
            if (i >= 0 && i < (NSInteger)opts.count) d[cle] = opts[i];
        }
    }
    NSData *json = [NSJSONSerialization dataWithJSONObject:d options:0 error:nil];
    if (json != nil) {
        NSString *s = [[NSString alloc] initWithData:json encoding:NSUTF8StringEncoding];
        free(gResultat);
        gResultat = strdup(s.UTF8String);
    }
    [gFen close];
}

- (void)annuler:(id)sender {
    [gFen close];
}

- (void)windowWillClose:(NSNotification *)n {
    gFen = nil;
    gCtrl = gType = gVal = gPas = gOpts = nil;
}

@end

// hauteur d'un champ dans la pile de son onglet
static CGFloat hauteurChamp(NSDictionary *ch) {
    NSString *type = ch[@"type"];
    if ([type isEqualToString:@"curseur"] || [type isEqualToString:@"entier"]) {
        return [ch[@"aide"] length] > 0 ? 68 : 52;
    }
    return 34;
}

// construit un champ dans vue, son haut a yTop du sommet ; renvoie sa hauteur
static CGFloat construireChamp(NSView *vue, NSDictionary *ch, CGFloat yTop) {
    NSString *type = ch[@"type"], *cle = ch[@"cle"];
    CGFloat H = vue.frame.size.height, L = vue.frame.size.width - 24;
    CGFloat h = hauteurChamp(ch);
    CGFloat yHaut = H - yTop; // ordonnee Cocoa du sommet du champ

    NSTextField *nom = [NSTextField labelWithString:ch[@"nom"]];
    nom.frame = NSMakeRect(12, yHaut - 18, L - 70, 17);
    [vue addSubview:nom];

    if ([type isEqualToString:@"curseur"] || [type isEqualToString:@"entier"]) {
        double pas = [ch[@"pas"] doubleValue];
        double val = [ch[@"valeur"] doubleValue];
        NSTextField *lv = [NSTextField labelWithString:formatValeur(val, pas)];
        lv.frame = NSMakeRect(12 + L - 60, yHaut - 18, 60, 17);
        lv.alignment = NSTextAlignmentRight;
        lv.textColor = [NSColor controlAccentColor];
        lv.font = [NSFont monospacedDigitSystemFontOfSize:13 weight:NSFontWeightSemibold];
        [vue addSubview:lv];

        NSSlider *s = [NSSlider sliderWithValue:val
                                       minValue:[ch[@"min"] doubleValue]
                                       maxValue:[ch[@"max"] doubleValue]
                                         target:gCible
                                         action:@selector(curseurChange:)];
        s.frame = NSMakeRect(12, yHaut - 42, L, 22);
        s.identifier = cle;
        if ([type isEqualToString:@"entier"] || pas >= 1) {
            s.allowsTickMarkValuesOnly = YES;
            s.numberOfTickMarks = (NSInteger)(([ch[@"max"] doubleValue] - [ch[@"min"] doubleValue]) / (pas >= 1 ? pas : 1)) + 1;
        }
        [vue addSubview:s];
        gCtrl[cle] = s;
        gVal[cle] = lv;
        gPas[cle] = @(pas);

        if ([ch[@"aide"] length] > 0) {
            NSTextField *aide = [NSTextField labelWithString:ch[@"aide"]];
            aide.frame = NSMakeRect(12, yHaut - 60, L, 14);
            aide.font = [NSFont systemFontOfSize:11];
            aide.textColor = [NSColor secondaryLabelColor];
            [vue addSubview:aide];
        }
    } else if ([type isEqualToString:@"choix"]) {
        NSPopUpButton *p = [[NSPopUpButton alloc]
            initWithFrame:NSMakeRect(12 + L - 170, yHaut - 24, 170, 25) pullsDown:NO];
        NSArray *libelles = ch[@"libelles"], *options = ch[@"options"];
        [p addItemsWithTitles:libelles];
        NSUInteger idx = [options indexOfObject:ch[@"texte"]];
        if (idx != NSNotFound) [p selectItemAtIndex:(NSInteger)idx];
        [vue addSubview:p];
        gCtrl[cle] = p;
        gOpts[cle] = options;
    } else if ([type isEqualToString:@"case"]) {
        NSButton *b = [NSButton checkboxWithTitle:@"" target:nil action:nil];
        b.state = [ch[@"coche"] boolValue] ? NSControlStateValueOn : NSControlStateValueOff;
        b.frame = NSMakeRect(12 + L - 24, yHaut - 21, 24, 20);
        [vue addSubview:b];
        gCtrl[cle] = b;
    }
    return h;
}

void dialogue_ouvrir(const char *desc) {
    @autoreleasepool {
        if (gFen != nil) { // deja ouverte : au premier plan
            [NSApp activateIgnoringOtherApps:YES];
            [gFen makeKeyAndOrderFront:nil];
            return;
        }
        NSData *data = [NSData dataWithBytes:desc length:strlen(desc)];
        NSDictionary *d = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
        if (d == nil) return;

        if (gCible == nil) gCible = [DialogueSinge new];
        gCtrl = [NSMutableDictionary new];
        gType = [NSMutableDictionary new];
        gVal = [NSMutableDictionary new];
        gPas = [NSMutableDictionary new];
        gOpts = [NSMutableDictionary new];

        NSArray *onglets = d[@"onglets"];
        // hauteur du plus grand onglet
        CGFloat contenuH = 0;
        for (NSDictionary *o in onglets) {
            CGFloat h = 16;
            for (NSDictionary *ch in o[@"champs"]) h += hauteurChamp(ch) + 8;
            if (h > contenuH) contenuH = h;
        }

        CGFloat largFen = 470, tabH = contenuH + 46, fenH = tabH + 72;
        gFen = [[NSWindow alloc]
            initWithContentRect:NSMakeRect(0, 0, largFen, fenH)
                      styleMask:NSWindowStyleMaskTitled | NSWindowStyleMaskClosable
                        backing:NSBackingStoreBuffered
                          defer:NO];
        gFen.title = d[@"titre"];
        gFen.releasedWhenClosed = NO;
        gFen.delegate = gCible;

        NSTabView *tabs = [[NSTabView alloc]
            initWithFrame:NSMakeRect(10, 54, largFen - 20, tabH)];
        for (NSDictionary *o in onglets) {
            NSTabViewItem *item = [[NSTabViewItem alloc] initWithIdentifier:o[@"titre"]];
            item.label = o[@"titre"];
            NSView *v = [[NSView alloc]
                initWithFrame:NSMakeRect(0, 0, largFen - 40, contenuH)];
            CGFloat yTop = 12;
            for (NSDictionary *ch in o[@"champs"]) {
                gType[ch[@"cle"]] = ch[@"type"];
                yTop += construireChamp(v, ch, yTop) + 8;
            }
            item.view = v;
            [tabs addTabViewItem:item];
        }
        [gFen.contentView addSubview:tabs];

        NSButton *ok = [NSButton buttonWithTitle:d[@"enregistrer"]
                                          target:gCible action:@selector(enregistrer:)];
        ok.frame = NSMakeRect(largFen - 170, 12, 158, 32);
        ok.keyEquivalent = @"\r";
        [gFen.contentView addSubview:ok];

        NSButton *non = [NSButton buttonWithTitle:d[@"annuler"]
                                           target:gCible action:@selector(annuler:)];
        non.frame = NSMakeRect(largFen - 268, 12, 92, 32);
        non.keyEquivalent = @"\033";
        [gFen.contentView addSubview:non];

        [gFen center];
        [NSApp activateIgnoringOtherApps:YES];
        [gFen makeKeyAndOrderFront:nil];
    }
}

char *dialogue_resultat(void) {
    char *r = gResultat;
    gResultat = NULL;
    return r;
}
