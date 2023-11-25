let w, h = @Dimensions()

let d = if color == #00000000{
    (w / 5)
}  else {
    (w / 8)
}
let size = w - d * 2

let bh = size + 158
let offsety = (h - bh) / 2

if has_bg {
    @DrawImageCovered(bg_image, 0, 0, w, h)
} else {
    @Clear(color)
}


with @PushRelative(d, offsety, size, size) {
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

    // @SetFont("discord-bold")
    // @SetFontSize(21)
    // @DrawString("Spotify", x + 23, y + 125)
    @SetColor(#ffffff70)
    // @Fill()

    @DrawSVGPath(spotify_logo)
    @Translate(x, y + 122)
    @Scale(0.13, 0.13)

    @Fill()
}