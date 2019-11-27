import React, { Fragment, useContext, useMemo } from "react";
import AppBarOrg from "@material-ui/core/AppBar";
import Toolbar from "@material-ui/core/Toolbar";
import Typography from "@material-ui/core/Typography";
import Avatar from "@material-ui/core/Avatar";
import { Theme } from "@material-ui/core/styles/createMuiTheme";
import useTheme from "@material-ui/core/styles/useTheme";
import useMediaQuery from "@material-ui/core/useMediaQuery";
import makeStyles from "@material-ui/styles/makeStyles";
import createStyles from "@material-ui/styles/createStyles";
import player from "static/player.png";
import { ComponentProperty } from "interfaces/component";
import { hueToRgb } from "interfaces/gamemap";
import LoginContext from "common/auth";
// import AdminPageContext from "common/admin";
// import SignIn from "./SignIn";
// import UserSettings from "./UserSettings";
import LogOut from "./LogOut";

// const Administrator = lazy(() => import("./Administrator"));

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      position: "absolute"
    },
    item: {
      marginRight: theme.spacing(1)
    },
    profile: {
      borderStyle: "solid",
      borderWidth: "2px",
      marginRight: theme.spacing(1),
      [theme.breakpoints.up("sm")]: {
        borderWidth: "4px",
        width: 50,
        height: 50
      }
    },
    name: {
      [theme.breakpoints.down("xs")]: {
        display: "none"
      }
    },
    grow: {
      flexGrow: 1
    },
    setting: {
      width: 50,
      height: 50,
      [theme.breakpoints.up("sm")]: {
        borderWidth: "4px",
        width: 50,
        height: 50
      }
    },
    inOperation: {
      color: theme.palette.error.main
    }
  })
);

interface UserInfoProperty extends ComponentProperty {
  myColor: string;
  myImage: string;
  myName: string;
}

const ShowUserInfo = (props: UserInfoProperty) => {
  const theme = useTheme();
  const classes = useStyles(theme);
  //const isAdmin = useContext(AdminPageContext);
  return (
    <Fragment>
      <Avatar
        style={{ borderColor: props.myColor }}
        className={classes.profile}
        src={props.myImage}
      />
      <div className={classes.name}>{props.myName}</div>
      <div className={classes.grow} />
      {/* <UserSettings /> */}

      {/* {isAdmin && (
        <Suspense fallback={<div>…</div>}>
          <Administrator />
        </Suspense>
      )} */}
      <LogOut />
    </Fragment>
  );
};

const LogInButton = () => {
  const theme = useTheme();
  const classes = useStyles(theme);
  return (
    <Fragment>
      <div className={classes.grow} />
      {/* <SignIn /> */}
    </Fragment>
  );
};

export default function() {
  const theme = useTheme();
  const classes = useStyles(theme);
  const isTiny = useMediaQuery(theme.breakpoints.down("xs"));
  const [[, my]] = useContext(LoginContext);
  const myColor = useMemo(() => {
    console.info("memo myColor");
    return my ? `rgb(${hueToRgb(my.hue).join(",")})` : "inherit";
  }, [my]);
  const myImage = my ? my.image : player;
  const myName = my ? my.name : "名無しさん";

  return (
    <AppBarOrg position="sticky" className={classes.root}>
      <Toolbar>
        <Typography className={classes.item} variant={isTiny ? "h6" : "h4"}>
          RushHour
        </Typography>

        {my ? (
          <ShowUserInfo myColor={myColor} myImage={myImage} myName={myName} />
        ) : (
          <LogInButton />
        )}
      </Toolbar>
    </AppBarOrg>
  );
}
