import argparse
from datetime import datetime, timedelta
import sys
import warnings

import numpy


if sys.version_info[0] < 3:
    # OK, ancients, I will support Python 2, but you owe me a beer
    input = raw_input


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--output", default="",
                        help="Path to the output file (empty for display).")
    parser.add_argument("--text-size", default=12,
                        help="Size of the labels and legend.")
    parser.add_argument("--backend", help="Matplotlib backend to use.")
    parser.add_argument("--style", choices=["black", "white"], default="black",
                        help="Plot's general color scheme.")
    parser.add_argument("--relative", action="store_true",
                        help="Occupy 100% height for every measurement.")
    parser.add_argument(
        "--resample", default="year",
        help="The way to resample the time series. Possible values are: "
             "\"month\", \"year\", \"no\", \"raw\" and pandas offset aliases ("
             "http://pandas.pydata.org/pandas-docs/stable/timeseries.html"
             "#offset-aliases).")
    args = parser.parse_args()
    return args


def main():
    args = parse_args()

    import matplotlib
    if args.backend:
        matplotlib.use(args.backend)
    import matplotlib.pyplot as pyplot
    import pandas

    start, granularity, sampling = input().split()
    start = datetime.fromtimestamp(int(start))
    granularity = int(granularity)
    sampling = int(sampling)
    matrix = numpy.array([numpy.fromstring(line, dtype=int, sep=" ")
                          for line in sys.stdin.read().split("\n")[:-1]]).T
    date_range_sampling = pandas.date_range(
        start + timedelta(days=sampling), periods=matrix.shape[1],
        freq="%dD" % sampling)
    if args.resample not in ("no", "raw"):
        aliases = {
            "year": "A",
            "month": "M"
        }
        args.resample = aliases.get(args.resample, args.resample)
        daily_matrix = numpy.zeros(
            (matrix.shape[0] * granularity, matrix.shape[1]),
            dtype=numpy.float32)
        daily_start = 1 if "M" in args.resample else 0
        for i in range(daily_start, matrix.shape[0]):
            daily_matrix[i * granularity:(i + 1) * granularity] = \
                matrix[i] / granularity
        date_range_granularity = pandas.date_range(
            start, periods=daily_matrix.shape[0], freq="1D")
        df = pandas.DataFrame({
            dr: pandas.Series(row, index=date_range_sampling)
            for dr, row in zip(date_range_granularity, daily_matrix)
        }).T
        df = df.resample(args.resample).sum()
        if "M" in args.resample:
            row0 = matrix[0]
        matrix = df.as_matrix()
        if "M" in args.resample:
            matrix[0] = row0
            for i in range(1, min(*matrix.shape)):
                matrix[i, i] += matrix[i, :i].sum()
                matrix[i, :i] = 0
        if args.resample in ("year", "A"):
            labels = [dt.year for dt in df.index]
        elif args.resample in ("month", "M"):
            labels = [dt.strftime("%Y %B") for dt in df.index]
        else:
            labels = [dt.date() for dt in df.index]
    else:
        labels = [
            "%s - %s" % ((start + timedelta(days=i * granularity)).date(),
                         (start + timedelta(days=(i + 1) * granularity)).date())
            for i in range(matrix.shape[0])]
        if len(labels) > 18:
            warnings.warn("Too many labels - consider resampling.")
        args.resample = "M"
    if args.style == "white":
        pyplot.gca().spines["bottom"].set_color("white")
        pyplot.gca().spines["top"].set_color("white")
        pyplot.gca().spines["left"].set_color("white")
        pyplot.gca().spines["right"].set_color("white")
        pyplot.gca().xaxis.label.set_color("white")
        pyplot.gca().yaxis.label.set_color("white")
        pyplot.gca().tick_params(axis="x", colors="white")
        pyplot.gca().tick_params(axis="y", colors="white")
    if args.relative:
        for i in range(matrix.shape[1]):
            matrix[:, i] /= matrix[:, i].sum()
        pyplot.ylim(0, 1)
        legend_loc = 3
    else:
        legend_loc = 2
    pyplot.stackplot(date_range_sampling, matrix, labels=labels)
    legend = pyplot.legend(loc=legend_loc, fontsize=args.text_size)
    frame = legend.get_frame()
    frame.set_facecolor("black" if args.style == "white" else "white")
    frame.set_edgecolor("black" if args.style == "white" else "white")
    for text in legend.get_texts():
        text.set_color(args.style)
    pyplot.ylabel("Lines of code", fontsize=args.text_size)
    pyplot.xlabel("Time", fontsize=args.text_size)
    pyplot.tick_params(labelsize=args.text_size)
    pyplot.xlim(date_range_sampling[0], date_range_sampling[-1])
    pyplot.gcf().set_size_inches(12, 9)
    locator = pyplot.gca().xaxis.get_major_locator()
    # set the optimal xticks locator
    if "M" not in args.resample:
        pyplot.gca().xaxis.set_major_locator(matplotlib.dates.YearLocator())
    locs = pyplot.gca().get_xticks().tolist()
    if len(locs) >= 16:
        pyplot.gca().xaxis.set_major_locator(matplotlib.dates.YearLocator())
        locs = pyplot.gca().get_xticks().tolist()
        if len(locs) >= 16:
            pyplot.gca().xaxis.set_major_locator(locator)
    if locs[0] < pyplot.xlim()[0]:
        del locs[0]
    endindex = -1
    if len(locs) >= 2 and \
            pyplot.xlim()[1] - locs[-1] >= (locs[-1] - locs[-2]) / 2:
        locs.append(pyplot.xlim()[1])
        endindex = len(locs) - 1
    startindex = -1
    if len(locs) >= 2 and \
            locs[0] - pyplot.xlim()[0] >= (locs[1] - locs[0]) / 2:
        locs.append(pyplot.xlim()[0])
        startindex = len(locs) - 1
    pyplot.gca().set_xticks(locs)
    # hacking time!
    labels = pyplot.gca().get_xticklabels()
    if startindex >= 0:
        if "M" in args.resample:
            labels[startindex].set_text(date_range_sampling[0].date())
            labels[startindex].set_text = lambda _: None
        labels[startindex].set_rotation(30)
        labels[startindex].set_ha("right")
    if endindex >= 0:
        if "M" in args.resample:
            labels[endindex].set_text(date_range_sampling[-1].date())
            labels[endindex].set_text = lambda _: None
        labels[endindex].set_rotation(30)
        labels[endindex].set_ha("right")
    if not args.output:
        pyplot.gcf().canvas.set_window_title(
            "Hercules %d x %d (granularity %d, sampling %d)" %
            (matrix.shape + (granularity, sampling)))
        pyplot.show()
    else:
        pyplot.tight_layout()
        pyplot.savefig(args.output, transparent=True)

if __name__ == "__main__":
    sys.exit(main())
