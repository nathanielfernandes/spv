let w, h = @Dimensions()

let d = (w / 8)
let size = w - d * 2

@Clear(color)

with @PushRelative(d, d, size, size) {
    let iw, ih = @Dimensions()
    

    @DrawRoundedRectangle(0, 0, iw, ih + 158, 10.0)
    @SetColor(@rgba(0, 0, 0, 0.5))
    @Fill()


    let x = 20
    let x2 = x * 2
    @DrawRoundedRectangle(x, x, iw - x2, ih - x2, 5.0)


    @SetFilter("good")
    @Clip()
    @DrawImageSized(art, x, x, iw - x2, ih - x2)
    @ResetClip()


    let track_name = @truncate(track_name, 58, "...")
    let artist_name = @truncate(artist_name, 34, "...")

    @SetFont("discord-semibold")

    let y = size 

    @SetLineSpacing(0.8)
    @DrawTextBoxWrapped(track_name, x, y, 0.0, 0.0, iw - x2, 80, "left")
    @SetColor(#ffffff)
    @Fill()

    @SetFontSize(29)
    @DrawString(artist_name, x, y + 80)
    @SetColor(#ffffffab)
    @Fill()

    @SetFont("discord-bold")
    @SetFontSize(21)
    @DrawString("Spotify", x + 23, y + 125)
    @Fill()

    @DrawSVGPath(spotify_logo)
    @Translate(x, y + 125)
    @Scale(0.04, 0.04)

    @Fill()
}