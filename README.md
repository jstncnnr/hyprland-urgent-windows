# hyprland-urgent-windows
This script will add the `urgent` tag on windows when they receive an urgent
event, and will remove the tag when the window receives focus or is closed.

You can style these windows by matching on the tag. The below example adds
a red border to urgent windows.

```
windowrule = bordercolor rgba(ff0000aa), tag:urgent
```