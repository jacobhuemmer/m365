package cli

import (
	"flag"

	"github.com/masonhuemmer/m365/internal/adapters/fs"
	"github.com/masonhuemmer/m365/internal/app/files"
	"github.com/masonhuemmer/m365/internal/domain"
)

func runFiles(args []string, d Deps, format string) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeHelp(d.Stdout, filesHelp)
	}
	if d.Files == nil {
		return fail(d, domain.Usage("files not available"))
	}
	verb, args := args[0], args[1:]
	sess, err := session(d)
	if err != nil {
		return fail(d, err)
	}
	switch verb {
	case "root":
		r, err := files.Root(ctx(), d.Files, sess)
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, r)
	case "list":
		if hasHelp(args) {
			return writeHelp(d.Stdout, filesHelp)
		}
		fsset := flag.NewFlagSet("files list", flag.ContinueOnError)
		fsset.SetOutput(d.Stderr)
		folder := fsset.String("folder", "", "")
		top := fsset.Int("top", 0, "")
		page := fsset.String("page-token", "", "")
		if err := parseMixed(fsset, args); err != nil {
			return fail(d, domain.Usage(err.Error()))
		}
		p, err := files.List(ctx(), d.Files, sess, files.ListQuery{Folder: *folder, Top: *top, PageToken: *page})
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, p)
	case "get":
		if len(args) < 1 {
			return fail(d, domain.Usage("item id is required"))
		}
		it, err := files.Get(ctx(), d.Files, sess, args[0])
		if err != nil {
			return fail(d, err)
		}
		return success(d, format, it)
	case "download":
		return filesDownload(args, d, sess, format)
	case "upload":
		return filesUpload(args, d, sess, format)
	case "create-folder":
		return filesMkdir(args, d, sess, format)
	case "delete":
		return filesDelete(args, d, sess, format)
	case "move":
		return filesMove(args, d, sess, format)
	default:
		return fail(d, domain.Usagef("unknown files verb %q", verb))
	}
}

func filesDownload(args []string, d Deps, sess domain.Session, format string) int {
	if len(args) < 1 {
		return fail(d, domain.Usage("item id is required"))
	}
	id := args[0]
	fsset := flag.NewFlagSet("files download", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	out := fsset.String("out", "", "")
	ow := fsset.Bool("overwrite", false, "")
	if err := parseMixed(fsset, args[1:]); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	write := d.Write
	if write == nil {
		write = fs.WriteFile
	}
	path, err := files.Download(ctx(), d.Files, sess, id, *out, *ow, write)
	if err != nil {
		return fail(d, err)
	}
	return success(d, format, map[string]any{"path": path})
}

func filesUpload(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("files upload", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	file := fsset.String("file", "", "")
	folder := fsset.String("folder", "", "")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	of, err := fs.ReadAttach(*file)
	if err != nil {
		return fail(d, err)
	}
	id, err := files.Upload(ctx(), d.Files, sess, files.UploadInput{Folder: *folder, File: of, DryRun: *dry})
	if err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, map[string]any{"dry_run": true, "name": of.Name, "size": of.Size})
	}
	return success(d, format, map[string]any{"id": id})
}

func filesMkdir(args []string, d Deps, sess domain.Session, format string) int {
	fsset := flag.NewFlagSet("files create-folder", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	name := fsset.String("name", "", "")
	folder := fsset.String("folder", "", "")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	id, err := files.CreateFolder(ctx(), d.Files, sess, files.FolderInput{Parent: *folder, Name: *name, DryRun: *dry})
	if err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, map[string]any{"dry_run": true, "name": *name})
	}
	return success(d, format, map[string]any{"id": id})
}

func filesDelete(args []string, d Deps, sess domain.Session, format string) int {
	if len(args) < 1 {
		return fail(d, domain.Usage("item id is required"))
	}
	id := args[0]
	fsset := flag.NewFlagSet("files delete", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	dry := fsset.Bool("dry-run", false, "")
	_ = parseMixed(fsset, args[1:])
	if err := files.Delete(ctx(), d.Files, sess, id, *dry); err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, map[string]any{"dry_run": true, "id": id})
	}
	return success(d, format, map[string]any{"id": id, "deleted": true})
}

func filesMove(args []string, d Deps, sess domain.Session, format string) int {
	if len(args) < 1 {
		return fail(d, domain.Usage("item id is required"))
	}
	id := args[0]
	fsset := flag.NewFlagSet("files move", flag.ContinueOnError)
	fsset.SetOutput(d.Stderr)
	folder := fsset.String("folder", "", "")
	name := fsset.String("name", "", "")
	dry := fsset.Bool("dry-run", false, "")
	if err := parseMixed(fsset, args[1:]); err != nil {
		return fail(d, domain.Usage(err.Error()))
	}
	if err := files.Move(ctx(), d.Files, sess, files.MoveInput{ID: id, Folder: *folder, Name: *name, DryRun: *dry}); err != nil {
		return fail(d, err)
	}
	if *dry {
		return success(d, format, map[string]any{"dry_run": true, "id": id})
	}
	return success(d, format, map[string]any{"id": id})
}
