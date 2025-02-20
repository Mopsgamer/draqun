package controller_http

// Bind Cookie -> Header -> URI -> Query -> Form.
func (ctl ControllerHttp) BindAll(out any) error {
	if err := ctl.Ctx.Bind().Cookie(out); err != nil {
		return err
	}

	if err := ctl.Ctx.Bind().Header(out); err != nil {
		return err
	}

	if err := ctl.Ctx.Bind().URI(out); err != nil {
		return err
	}

	if err := ctl.Ctx.Bind().Query(out); err != nil {
		return err
	}

	if err := ctl.Ctx.Bind().Form(out); err != nil {
		return err
	}

	return nil
}
